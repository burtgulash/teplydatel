/// <reference path="jquery.d.ts" />


window.onload = function () {
    var conn;
    var statusBox = $("#status");

    var done = $("#done");
    var error = $("#error");
    var remaining = $("#remaining");

    var send_buf = [];
    var before_cursor = [];
    var error_arr = [];
    var after_cursor = remaining.text().split("");
    var length = after_cursor.length;

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
                conn.send("p " + send_buf.join(""));
                console.log("sending buf: ", send_buf.join(""));
                send_buf = [];
            }

            done.text(before_cursor.join(""));
            remaining.text(after_cursor.join(""));
        } else {
            error_arr.push(c);
            error.text(error_arr.join(""));
        }
    }

    // handle backspace
    function onkeydown(event) {
        if (event.which == 8) {
            error_arr.pop();
            error.text(error_arr.join(""));
        }
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
        conn.onmessage = function(event) {
            var data = event.data.split(" ");
            var cmd = data[0];

            var args = data.slice(1);
            if (cmd == "status") {
                status = args[0];
                if (status == "live") {
                    $(document).keypress(onkeypress).keydown(onkeydown);
                }
                statusBox.text(status);
            } else if (cmd == "r") {
                console.log(args);
            } else {
                console.log("unknown command", cmd);
            }

            console.log("message received: ", event.data);
        }
    }
}
