var Race = (function () {
    function Race(race_text) {
        this.race_text = race_text;
        this.yet = race_text.split(" ");
        this.sofar = [];
        this.in_progress = this.yet.shift();
        this.update_dom();
    }
    Race.prototype.on_type = function (event) {
        if (event.keyCode == 32) {
            var input = document.getElementById("sem-pis");
            console.log(input.value);
            console.log(this.in_progress);
            if (input.value.trim() == this.in_progress) {
                this.shift();
                input.value = "";
            }
            console.log(input.value);
            console.log(this.in_progress);
        }
    };
    Race.prototype.shift = function () {
        this.sofar.push(this.in_progress);
        this.in_progress = this.yet.shift();
        this.update_dom();
    };
    Race.prototype.update_dom = function () {
        document.getElementById("tos-napsals").textContent = this.sofar.join(" ");
        document.getElementById("tuto-pis").textContent = this.in_progress;
        document.getElementById("tuto-napises").textContent = this.yet.join(" ");
    };
    return Race;
})();
window.onload = function () {
    var race_text = document.getElementById("tuto-napises").textContent;
    var race = new Race(race_text);
    var input = document.getElementById("sem-pis");
    input.focus();
    input.addEventListener("keyup", function (event) {
        race.on_type(event);
    });
};
