function Tick(done, wpm) {
    this.done = done;
    this.wpm = wpm;
}

function Plot(container, width, height) {
    var margin = {
        top: 20,
        bottom: 20,
        left: 10,
        right: 10
    };

    this.progress = {};

    var max_wpm_estimate = 120;

    var svg = d3.select(container).append("svg")
                .attr("width", width)
                .attr("height", height);

    var x = d3.scale.linear()
            .domain([0, 100])
            .range([margin.left, width - margin.right]),
        y = d3.scale.linear()
            .domain([0, max_wpm_estimate])
            .range([height - margin.top, margin.bottom]);

    var lineFunction = d3.svg.line()
        .x(function(d) { return x(d.done); })
        .y(function(d) { return y(d.wpm); })
        .interpolate("linear");


    this.update = function(data) {
        var lines = svg.selectAll(".line")
            .attr("d", lineFunction)
            .data(data);

        lines.enter().append("path")
            .attr("class", "line")
            .attr("d", lineFunction)
            .attr("stroke", "blue")
            .attr("stroke-width", 4)
            .attr("fill", "none");
    };

    this.add_player = function(player_id) {
        console.log("adding player", player_id);
        this.progress[player_id] = [new Tick(0, 0)];
    };

    this.update_progress = function(player_id, done, wpm) {
        var progress = this.progress[player_id];
        if (!progress) {
            console.log("player does not exist",
                        {player_id: player_id});
            return;
        }

        progress.push(new Tick(done, wpm));

        this.update(d3.values(this.progress));
    };
}
