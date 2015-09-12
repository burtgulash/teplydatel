package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func (l *Lobby) lobby_handler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", struct{ Name string }{"John"})
}

func (l *Lobby) race_creator_handler(w http.ResponseWriter, r *http.Request) {
	race_type := r.URL.Query().Get("t")

	var race *Race
	if race_type == "verejny" {
		race = l.Create_or_join_public_race()
	} else if race_type == "soukromy" {
		race = l.Create_private_race()
	} else {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/zavod/"+race.Race_code, http.StatusFound)
}

func (l *Lobby) race_handler(w http.ResponseWriter, r *http.Request) {
	race_code := r.URL.Query().Get(":race_code")
	race, exists := l.races[race_code]
	if !exists {
		http.Error(w, "Race not found", http.StatusNotFound)
		return
	}

	templates.ExecuteTemplate(w, "race.html", race)
}

func (l *Lobby) ws_handler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	race_code := r.URL.Query().Get(":race_code")

	race, in := l.races[race_code]
	if !in {
		ws.WriteMessage(websocket.CloseMessage, []byte("Race does not exist"))
		ws.Close()
		return
	}

	// TODO authorize player to this race. Check if race is still waiting for players
	// else reject the connection
	player := &Player{
		name: "Jarda",
	}

	if err := race.join(player, ws); err != nil {
		ws.WriteMessage(websocket.CloseMessage, []byte("Forbidden to join the race"))
		ws.Close()
		return
	}
}
