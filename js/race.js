window.onload = function () {
    var conn;
    var statusBox = document.getElementById("status");
    if (window["WebSocket"]) {
        var parser = document.createElement("a");
        parser.href = window.location.href;
        var path = parser.pathname.split("/");
        var race_code = path[path.length - 1];
        var ws_uri = "ws://" + parser.host + "/ws/" + race_code;
        statusBox.textContent = "connecting...";
        console.log("connecting to", ws_uri);
        conn = new WebSocket(ws_uri);
        conn.onopen = function (event) {
            statusBox.textContent = "connection established";
        };
        conn.onclose = function (event) {
            statusBox.textContent = "connection closed";
            console.log("connection closed");
        };
        conn.onmessage = function (event) {
            console.log("message received");
        };
    }
};
