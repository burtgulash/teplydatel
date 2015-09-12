class Race {
    private sofar: string[];
    private in_progress: string;
    private yet: string[];

    private length: number;
    private done: number;

    constructor(public race_text) {
        // Remove double spaces just in case
        race_text = race_text.replace(" +", " ");

        this.length = race_text.length;
        this.done = 0;

        this.yet = race_text.split(" ");
        this.sofar = [];
        this.in_progress = this.yet.shift();

        this.update_dom();
    }

    public on_type(event) {
        var input = <HTMLInputElement>document.getElementById("sem-pis");

        // If last character
        if (this.done + this.in_progress.length >= this.length - 1) {
            if (input.value == this.in_progress) {
                input.readOnly = true;
                input.value = "FINISH!!";
            }
        } else if (event.keyCode == 32) {
            // .trim() must be used because spacebar can be held down and this
            // event will be fired only after it is released
            if (input.value.trim() == this.in_progress) {
                this.shift();

                this.done += this.in_progress.length + 1;
                input.value = "";
        console.log("LENGTH ", this.length, " ", this.done, " ", this.done + this.in_progress.length);
            }
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

