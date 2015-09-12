package main

import (
	"log"
	"time"
)

type Race struct {
	race_id      int64
	Race_code    string
	status       string
	created_time time.Time
	start_time   time.Time
	Race_text    *string
	lobby        *Lobby

	players    map[*Player]bool
	receive    chan []byte
	register   chan *Player
	unregister chan *Player
}

func NewRace(lobby *Lobby, race_code string) *Race {
	return &Race{
		lobby:      lobby,
		Race_code:  race_code,
		players:    make(map[*Player]bool),
		receive:    make(chan []byte),
		register:   make(chan *Player),
		unregister: make(chan *Player),
	}
}

func (r *Race) run() {
	timer := time.NewTimer(10 * time.Minute)
	defer func() {
		timer.Stop()
		r.lobby.unregister_race <- r
	}()

	for {
		select {
		case <-timer.C:
			break
		case player := <-r.register:
			r.players[player] = true
			log.Println("Player", player.name, "joined race", r.Race_code)

			// TODO after countdouwn is initiated, reset the timer
			// timer.Reset(5 * time.Minute)
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
