/// <reference path="jquery.d.ts" />


window.onload = function () {
    var conn;
    var notifyTimeout;
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

    function notifyProgress() {
        clearTimeout(notifyTimeout);

        conn.send("p " + error_counter +
                " " + send_buf.join(""));
        send_buf = [];
        error_counter = 0;

        if (after_cursor.length > 0)
            notifyTimeout = setTimeout(notifyProgress, 3 * 1000);
    }

    function onkeypress(event) {
        var expected = after_cursor[0];
        var c = String.fromCharCode(event.which);

        if (c == expected && error_arr.length == 0) {
            after_cursor.shift();
            before_cursor.push(c);

            // Each 5 successful characters or when 'remaining'
            // buffer depleted, send progress report
            send_buf.push(c);
            if (send_buf.length >= 5 || after_cursor.length == 0)
                notifyProgress();

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
        var standings = [];

        for (var player_id in race.players) {
            var player_status = race.players[player_id];
            var progress = player_status.progress;
            var pg;
            if (progress.done == race.len)
                pg = "hotovo!";
            else {
                var p = Math.round(100 * progress.done / race.len);
                pg = p + "%";
            }

            var error_rate = progress.errors/(progress.done + 1);
            error_rate = Math.min(1, error_rate);
            var accuracy = Math.round(100 * (1 - error_rate));

            standings.push({
                accuracy: accuracy,
                wpm: progress.wpm,
                player_id: player_id,
                finished: player_status.finished,
                rank: player_status.rank,
                progress: ""+pg,
                connected: player_status.connected
            });
        }

        standings.sort(function(a, b) {
            if (!a.finished && !b.finished)
                return a.player_id.localeCompare(b.player_id);
            if (!a.finished)
                return 1;
            if (!b.finished)
                return -1;
            return a.rank - b.rank;
        });

        var standings_elem = $("ul#standings");
        standings_elem.html("");
        for (var i = 0; i < standings.length; i++) {
            var s = standings[i];
            var text = "<li>";
            text += "Hráč " + s.player_id + ": ";
            text += s.progress;
            text += ", přesnost: " + s.accuracy + "%";
            text += ", wpm: " + s.wpm;

            if (s.finished)
                text += ", pořadí: " + s.rank + ".";
            else if (!s.connected)
                text += ", nedokončil";
            text += "</li>";

            standings_elem.append(text);
        }
    }

    function start_race() {
        $(document).keypress(onkeypress).keydown(onkeydown);
        statusBox.text("Piš!");
        notifyTimeout = setTimeout(notifyProgress, 3 * 1000);
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
            if (race.status == "live")
                start_race();
        } else if (cmd == "j") {
            race.players[player_id] = {
                id: player_id,
                name: args[0],
                connected: true,
                finished: false,
                rank: null,
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
        } else if (cmd == "c") {
            statusBox.text(+args[0] + "s zbývá...");
        } else if (cmd == "f") {
            player.finished = true;
            player.rank = +args[0];
        } else if (cmd == "d") {
            player.connected = false;
        } else {
            console.log("unknown command", cmd);
        }

        updateStandings();
        // DEBUG
        //console.log("message received: ", event.data);
    }

    if (window["WebSocket"]) {
        $(document).keydown(function(event) {
            if (event.which == 8)
                event.preventDefault();
        });
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
