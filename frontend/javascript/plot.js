function Plot(container, width, height) {
    this.margin = {
        top: 20,
        bottom: 20,
        left: 10,
        right: 10
    };

    this.progress = {};

    this.width = width - this.margin.left - this.margin.right;
    this.height = height - this.margin.top - this.margin.bottom;

    this.svg = d3.select(container).append("svg")
                 .attr("width", width)
                 .attr("height", height);

    this.x = d3.scale.linear()
               .range([0, this.width]);

    this.y = d3.scale.linear()
               .range([this.height, 0]);

    this.svg.append("text")
        .text("TEST");
}

function Tick(done, wpm) {
    this.done = done;
    this.wpm = wpm;
}

Plot.prototype.add_player = function(player_id) {
    this.progress[player_id] = [Tick(0, 0)];
};

Plot.prototype.update_progress = function(player_id, done, wpm) {
    var progress = this.progress[player_id];
    if (!progress) {
        console.log("player does not exist",
                    {player_id: player_id});
        return;
    }

    var last_tick = progress[progress.length - 1];
    var current_tick = Tick(done, wpm);
    progress.push(current_tick);
    // TODO this.lineto(last_tick, current_tick)
};

Plot.prototype.redraw = function() {
// update_progress for all players? and progresses recorded fo each player
};
