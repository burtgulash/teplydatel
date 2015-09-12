package main

import (
	"log"
)

type Lobby struct {
	players map[*Player]bool
	races   map[Race_code]*Race

	create_race     chan chan *Race
	unregister_race chan *Race
}

func NewLobby() *Lobby {
	return &Lobby{
		players: make(map[*Player]bool),
		races:   make(map[Race_code]*Race),

		create_race:     make(chan chan *Race),
		unregister_race: make(chan *Race),
	}
}

func (l *Lobby) run() {
	log.Println("Lobby running")
	for {
		select {
		case race_request_req := <-l.create_race:
			race := l.make_race()
			l.races[race.race_code] = race

			// Run the race!
			go race.run()
			log.Println("Created race", race.race_code)

			// Send the newly created race back to requester
			race_request_req <- race
			close(race_request_req)
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

func (l *Lobby) Create_private_race() *Race {
	req := make(chan *Race, 1)
	l.create_race <- req
	race := <-req
	return NewRace(l, race.race_code)
}

func (l *Lobby) Create_or_join_public_race() *Race {
	return l.Create_private_race()
}
