package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type RaceMessage struct {
	conn *connection
	data string
}

type PlayerProgress struct {
	conn   *connection
	player *Player
	race   *Race
	done   int
}

type Race struct {
	race_id      int64
	Race_code    string
	status       string
	created_time *time.Time
	start_time   *time.Time
	Race_text    *string
	lobby        *Lobby

	players map[*connection]*PlayerProgress

	receive      chan RaceMessage
	player_join  chan *PlayerProgress
	player_leave chan *Player
}

func NewRace(lobby *Lobby, race_code string) *Race {
	return &Race{
		lobby:        lobby,
		Race_code:    race_code,
		players:      make(map[*connection]*PlayerProgress),
		receive:      make(chan RaceMessage, 16),
		player_join:  make(chan *PlayerProgress, 2),
		player_leave: make(chan *Player, 2),
	}
}

func (r *Race) set_status(status string) {
	old_status := r.status
	r.status = status
	r.broadcast("status " + status)
	log.Printf("INFO race %s changed status %s -> %s", r.Race_code, old_status, status)
}

func (r *Race) run() {
	countdown := make(chan bool, 1)

	r.set_status("created")
	for {
		select {

		case <-countdown:
			log.Printf("INFO race %s started!", r.Race_code)
			r.set_status("live")

		case msg := <-r.receive:
			if msg.data[0] == 'p' {
				sender := r.players[msg.conn].player

				m := msg.data[2:]
				b := fmt.Sprintf("r %d %s", sender.player_id, m)
				log.Println("DEBUG broadcasting message: " + b)
				r.broadcast(b)
			}

		case pp := <-r.player_join:
			r.players[pp.conn] = pp
			log.Printf("INFO player %s joined race %s", pp.player.name, r.Race_code)

			if len(r.players) <= 1 {
			} else if r.start_time == nil {
				countdown_period := time.Second * 10
				start_time := time.Now().Add(countdown_period)
				r.start_time = &start_time

				log.Printf("INFO race %s set start time to %s", r.Race_code, r.start_time.Format("15:04:05.000"))
				go func() {
					<-time.After(start_time.Sub(time.Now()))
					countdown <- true
				}()
			} else {
				still_time_to_join := time.Second * 4
				time.Now().Before(r.start_time.Add(-still_time_to_join))
			}

		case player := <-r.player_leave:
			for _, pp := range r.players {
				if pp.player == player {
					pp.conn.close()
					log.Printf("INFO player %s left race %s", player.name, r.Race_code)
					break
				}
			}
		}
	}
}

func (r *Race) broadcast(message string) {
	for conn := range r.players {
		conn.send <- message
	}
}

func (r *Race) join(player *Player, ws *websocket.Conn) (*connection, error) {
	conn := NewConnection(ws, player, r.receive)
	pp := PlayerProgress{
		conn:   conn,
		player: player,
		race:   r,
		done:   0,
	}

	r.player_join <- &pp
	return conn, nil
}

func (r *Race) leave(player *Player) error {
	r.player_leave <- player
	return nil
}
