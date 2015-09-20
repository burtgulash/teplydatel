/// <reference path="jquery.d.ts" />


window.onload = function () {
    var conn;
    var race = {
        status: "created",
        len: 0,
        players: {}
    };
    var statusBox = $("#status");

    var done = $("#done");
    var error = $("#error");
    var remaining = $("#remaining");

    var send_buf = [];
    var error_counter = 0;
    var before_cursor = [];
    var error_arr = [];
    var after_cursor = remaining.text().split("");
    race.len = after_cursor.length;

    function onkeypress(event) {
        var expected = after_cursor[0];
        var c = String.fromCharCode(event.which);

        if (c == expected && error_arr.length == 0) {
            after_cursor.shift();
            before_cursor.push(c);

            // Each 5 successful characters or when 'remaining'
            // buffer depleted, send progress report
            send_buf.push(c);
            if (send_buf.length >= 5 || after_cursor.length == 0) {
                conn.send("p " + error_counter +
                        " " + send_buf.join(""));

                send_buf = [];
                error_counter = 0;
            }

            done.text(before_cursor.join(""));
            remaining.text(after_cursor.join(""));
        } else if (error_arr.length + 1 <= after_cursor.length) {
            error_counter ++;

            if (c == " ")
                c = "_";

            error_arr.push(c);
            error.text(error_arr.join(""));
            remaining.text(after_cursor.slice(error_arr.length).join(""));
        }
    }

    // handle backspace
    function onkeydown(event) {
        if (event.which == 8) {
            error_arr.pop();
            error.text(error_arr.join(""));
            remaining.text(after_cursor.slice(error_arr.length).join(""));
        }
    }

    function updateStandings() {
        var standings = $("ul#standings");
        standings.html("");
        for (var player_id in race.players) {
            var progress = race.players[player_id].progress;
            var pg;
            if (progress.done == race.len)
                pg = "finished!";
            else {
                var p = Math.round(100 * progress.done / race.len);
                pg = p + "%";
            }

            var accuracy = Math.min(race.len, progress.errors) / (progress.done + 1);
            accuracy = Math.round(100 * (1 - accuracy));

            standings.append("<li>Player " + player_id +
                    ": " + pg +
                    ", accuracy: " + accuracy + "%" +
                    ", wpm: " + progress.wpm + 
                    "</li>");
        }
    }

    function onwsmessage(event) {
        var data = event.data.split(" ");
        var cmd = data[0];
        var player_id = data[1];
        var player;

        if (player_id == "glob") {
            player = null;
        } else if (player_id in race.players) {
            player = race.players[player_id];
        } else {
            console.log("Player", player_id, "not found!");
        }

        var args = data.slice(2);
        if (cmd == "s") {
            race.status = args[0];
            if (race.status == "live") {
                $(document).keypress(onkeypress).keydown(onkeydown);
            }
            statusBox.text(race.status);
            updateStandings();
        } else if (cmd == "j") {
            race.players[player_id] = {
                id: player_id,
                name: args[0],
                connected: true,
                finished: false,
                progress: {
                    done: 0,
                    errors: 0,
                    wpm: 0
                }
            };
        } else if (cmd == "r") {
            var progress = player.progress;
            progress.done = +args[0];
            progress.errors = +args[1];
            progress.wpm = +args[2];
            updateStandings();
        } else if (cmd == "c") {
            statusBox.text(+args[0] + "s remain...");
        } else if (cmd == "f") {
            player.finished = true;
            console.log("player", player_id, "finished!!");
        } else if (cmd == "d") {
            console.log("player", player_id, "disconnected");
            player.connected = false;
        } else {
            console.log("unknown command", cmd);
        }

        // DEBUG
        console.log("message received: ", event.data);
    }

    if (window["WebSocket"]) {
        var parser = document.createElement("a");
        parser.href = window.location.href;

        var path = parser.pathname.split("/");

        var race_code = path[path.length - 1];
        var ws_uri = "ws://" + parser.host + "/ws/" + race_code;
        statusBox.text("connecting...");
        console.log("connecting to", ws_uri);


        conn = new WebSocket(ws_uri);
        conn.onopen = function(event) {
            console.log("ws connection established");
            statusBox.text("ws connection established");
        }
        conn.onclose = function(event) {
            console.log("ws connection closed");
            statusBox.text("ws connection closed");
        }
        conn.onmessage = onwsmessage;
    }
}
