window.onload = function () {
    var conn;

    if (window["WebSocket"]) {
        var parser = document.createElement("a");
        parser.href = window.location.href;

        var path = parser.pathname.split("/");

        var race_code = path[path.length - 1];
        var ws_uri = "ws://" + parser.host + "/ws/" + race_code;
        console.log("URI", ws_uri);

        conn = new WebSocket(ws_uri);
        conn.onclose = function(event) {
            console.log("connection closed!");
        }
        conn.onmessage = function(event) {
            console.log("message received");
        }
    }
}
