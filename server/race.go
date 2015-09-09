package main

import (
	"time"
)

type Race struct {
	race_id      int64
	race_code    [7]byte
	status       string
	created_time time.Time
	start_time   time.Time
	race_text    string

	players    map[*Player]bool
	receive    chan []byte
	register   chan *Player
	unregister chan *Player
}

func NewRace(race_code string) *Race {
	return &Race{
		race_code:  race_code,
		players:    make(map[*Player]bool),
		receive:    make(chan []byte),
		register:   make(chan *Player),
		unregister: make(chan *Player),
	}
}

func (r *Race) run() {
	for {
		select {
		case player := <-r.register:
			r.players[player] = true
		case player := <-r.unregister:
			if _, ok := r.players[player]; ok {
				delete(r.players, player)
				close(player.conn.send)
			}
		case m := <-r.receive:
			// TODO parse and process message here
			for player := range r.players {
				select {
				case player.conn.send <- m:
				default:
					r.unregister <- player
				}
			}
		}
	}
}
