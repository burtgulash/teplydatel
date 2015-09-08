class Race {
    sofar: string[];
    in_progress: string;
    yet: string[];

    constructor(public race_text) {
        this.yet = race_text.split(" ");
        this.sofar = [];
        this.in_progress = this.yet.shift();

        this.update_dom();
    }

    public on_type(event) {
        if (event.keyCode == 32) {
            var input = <HTMLInputElement>document.getElementById("sem-pis");
            console.log(input.value);
            console.log(this.in_progress);
            if (input.value.trim() == this.in_progress) {
                this.shift();
                input.value = "";
            }
            console.log(input.value);
            console.log(this.in_progress);
        }
    }

    private shift() {
        this.sofar.push(this.in_progress);
        this.in_progress = this.yet.shift();
        this.update_dom();
    }

    private update_dom() {
        document.getElementById("tos-napsals").textContent = this.sofar.join(" ");
        document.getElementById("tuto-pis").textContent = this.in_progress;
        document.getElementById("tuto-napises").textContent = this.yet.join(" ");
    }

}

window.onload = function () {
    var race_text = document.getElementById("tuto-napises").textContent;
    var race = new Race(race_text);

    var input = document.getElementById("sem-pis");
    input.focus();
    input.addEventListener("keyup", (event) => { 
        race.on_type(event)
    });
}

