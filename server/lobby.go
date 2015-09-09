package main

type Lobby struct {
	players map[*Player]bool
	races   map[string]*Race

	register_race   chan *Race
	unregister_race chan *Race
}

func NewLobby() *Lobby {
	return &Lobby{
		players: make(map[*Player]bool),
		races:   make(map[string]*Race),

		register_race:   make(chan chan *Race),
		unregister_race: make(chan *Race),
	}
}

func (l *lobby) run() {
	for {
		select {
		case race_request := <-l.register_race:
			race := l.make_race()
			l.races[race_code] = race

			// Run the race!
			go race.run()

			// Send the newly created race back to requester
			race_request <- race
			close(race_request)
		case race := <-l.unregister_race:
			if _, in := l.races[race_code]; in {
				delete(l.races, race_code)
			} else {
				log.Println("ERROR", "can't unregister non-existing race", race_code)
			}
		}
	}
}

func (l *lobby) make_race() *Race {
	// TODO create race_code here
	race_code := "xxxxxxx"
	return NewRace(race_code)
}

func (l *lobby) ws_handler(w http.ResponseWriter, r *http.Request) {
	race_code := r.URL.Query().Get(":race_code")
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

	if ws, err := upgrader.Upgrade(w, r, nil); err != nil {
		log.Println(err)
		return
	}

	player.conn = &connection{
		send: make(chan []byte, 256),
		ws:   ws,
		race: race,
	}

	race.register <- player
}
