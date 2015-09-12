package main

import (
	"log"
	"math/rand"
)

const (
	letters        = "abcdefghijklmnopqrstuvwxyz0123456789"
	race_code_size = 7
)

type Lobby struct {
	players map[*Player]bool
	races   map[string]*Race

	create_race     chan chan *Race
	unregister_race chan *Race
}

func NewLobby() *Lobby {
	return &Lobby{
		players: make(map[*Player]bool),
		races:   make(map[string]*Race),

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

func gen_code(size int) string {
	b := make([]byte, size)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b[:])
}

func (l *Lobby) make_race() *Race {
	// TODO create race_code here
	var race_code string
	for {
		race_code = gen_code(race_code_size)
		if _, exists := l.races[race_code]; !exists {
			break
		}
	}
	return NewRace(l, race_code)
}

func (l *Lobby) Create_private_race() *Race {
	req := make(chan *Race, 1)
	l.create_race <- req
	race := <-req
	return race
}

func (l *Lobby) Create_or_join_public_race() *Race {
	return l.Create_private_race()
}
