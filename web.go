package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
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

	http.Redirect(w, r, "/zavod/"+string(race.race_code[:]), http.StatusFound)
}

func (l *Lobby) race_handler(w http.ResponseWriter, r *http.Request) {
	race_code := r.URL.Query().Get(":race_code")

	templates.ExecuteTemplate(w, "race.html", struct{ race_code string }{race_code})
}

func (l *Lobby) ws_handler(w http.ResponseWriter, r *http.Request) {
	race_code_arg := r.URL.Query().Get(":race_code")
	log.Println("cioa", race_code_arg)

	var race_code Race_code
	copy(race_code[:], race_code_arg)
	race, in := l.races[race_code]
	if !in {
		http.Error(w, "Race does not exist", http.StatusNotFound)
		return
	}

	// TODO authorize player to this race. Check if race is still waiting for players
	// else reject the connection
	player := &Player{
		name: "Jarda",
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	player.conn = &connection{
		send:   make(chan []byte, 256),
		ws:     ws,
		player: player,
		race:   race,
	}

	race.register <- player
	go player.conn.ws_reader()
	player.conn.ws_writer()
}
