package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Lobby struct {
	players map[*Player]bool
	races   map[Race_code]*Race

	register_race   chan chan *Race
	unregister_race chan *Race
}

func NewLobby() *Lobby {
	return &Lobby{
		players: make(map[*Player]bool),
		races:   make(map[Race_code]*Race),

		register_race:   make(chan chan *Race),
		unregister_race: make(chan *Race),
	}
}

func (l *Lobby) run() {
	for {
		select {
		case race_request := <-l.register_race:
			race := l.make_race()
			l.races[race.race_code] = race

			// Run the race!
			go race.run()

			// Send the newly created race back to requester
			race_request <- race
			close(race_request)
		case race := <-l.unregister_race:
			if _, in := l.races[race.race_code]; in {
				delete(l.races, race.race_code)
			} else {
				log.Println("ERROR", "can't unregister non-existing race", race.race_code)
			}
		}
	}
}

func (l *Lobby) make_race() *Race {
	// TODO create race_code here
	var race_code Race_code
	copy(race_code[:], "xxxxxxx")
	return NewRace(l, race_code)
}

func (l *Lobby) ws_handler(w http.ResponseWriter, r *http.Request) {
	race_code_arg := r.URL.Query().Get(":race_code")

	var race_code Race_code
	copy(race_code[:], race_code_arg)
	race, in := l.races[race_code]
	if !in {
		http.Error(w, "Race does not exist", 404)
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
