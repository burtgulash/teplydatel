/// <reference path="typings/jquery.d.ts" />
/// <reference path="typings/plot.d.ts" />


window.onload = function () {
    var conn;

    var notifyTimeout;
    var lastInput = Date.now();

    var race = {
        code: null,
        status: "created",
        race_type: null,
        is_started: false,
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

    var plot = new Plot("#chart", 720, 160);

    function notifyProgress() {
        clearTimeout(notifyTimeout);

        conn.send(JSON.stringify({
            typ: "progress",
            done: send_buf.join(""),
            errors: error_counter
        }));

        send_buf = [];
        error_counter = 0;

        // if there are still characters left to type and there
        // is some user activity in recent couple of seconds,
        // send progress to server
        if (after_cursor.length > 0 &&
                Date.now() - lastInput <= 10 * 1000)
            notifyTimeout = setTimeout(notifyProgress, 3 * 1000);
    }

    function onkeypress(event) {
        lastInput = Date.now();

        var expected = after_cursor[0];
        var c = String.fromCharCode(event.which);

        if (c == expected && error_arr.length == 0) {
            if (!race.is_started && race.race_type == "practice")
                start_practice_race();

            after_cursor.shift();
            before_cursor.push(c);

            // Each 5 successful characters or when 'remaining'
            // buffer depleted, send progress report
            send_buf.push(c);
            if (send_buf.length >= 3 || after_cursor.length == 0)
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
        lastInput = Date.now();

        if (event.which == 8) {
            error_arr.pop();
            error.text(error_arr.join(""));
            remaining.text(after_cursor.slice(error_arr.length).join(""));
        } else if (event.which == 32) {
            onkeypress(event);
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
            var accuracy = Math.floor(100 * (1 - error_rate));

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

    function allow_race() {
        $(".fields").focus();
        $(".fields").keypress(onkeypress).keydown(onkeydown);
        statusBox.text("Piš!");
    }

    function start_race() {
        race.is_started = true;
        console.log("RACE STARTED");
        notifyTimeout = setTimeout(notifyProgress, 3 * 1000);
    }

    function start_practice_race() {
        start_race();
        conn.send(JSON.stringify({
            typ: "start",
            at: (new Date()).toISOString()
        }));
    }

    function onwsmessage(event) {
        var data = JSON.parse(event.data);
        var typ = data.typ;
        var player_id = data.plid;
        var player;

        if (player_id == 0) {
            player = null;
        } else if (player_id in race.players) {
            player = race.players[player_id];
        } else {
            console.log("Player", player_id, "not found!");
        }

        if (typ == "status") {
            race.status = data.status;
            if (race.race_type != "practice" && race.status == "live") {
                allow_race();
                start_race();
            }
        } else if (typ == "info") {
            race.code = data.code;
            race.race_type = data.race_type;
            if (race.race_type == "practice") {
                allow_race();
            }
        } else if (typ == "joined") {
            var color = data.color;
            console.log("joined color", color);

            race.players[player_id] = {
                id: player_id,
                name: player_id,
                color: color,
                connected: true,
                finished: false,
                rank: null,
                progress: {
                    done: 0,
                    errors: 0,
                    wpm: 0
                }
            };

            plot.add_player(player_id, color);
        } else if (typ == "progress") {
            var progress = player.progress;
            progress.done = +data.done;
            progress.errors = +data.errors;
            progress.wpm = +data.wpm;

            var done_percent = progress.done * 100 / race.len;
            console.log("Done percent", done_percent);

            plot.update_progress(player_id,
                                 done_percent,
                                 progress.wpm);
        } else if (typ == "countdown") {
            statusBox.text(+data.remains + "s zbývá...");
        } else if (typ == "finished") {
            player.finished = true;
            player.rank = +data.rank;
        } else if (typ == "disconnected") {
            player.connected = false;
        } else {
            console.log("unknown command", typ);
        }

        updateStandings();
        // DEBUG
        console.log("message received: ", event.data);
    }

    if (window["WebSocket"]) {
        // disable default behaviour for some keys
        $(document).keydown(function(event) {
            // 8 = backspace
            // 32 = space
            if (event.which == 8 || event.which == 32)
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
