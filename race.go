package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
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
	lock    sync.Mutex

	receive   chan RaceMessage
	countdown chan bool
}

func NewRace(lobby *Lobby, race_code string) *Race {
	return &Race{
		lobby:     lobby,
		Race_code: race_code,
		players:   make(map[*connection]*PlayerProgress),
		receive:   make(chan RaceMessage, 16),
		countdown: make(chan bool, 1),
	}
}

func (r *Race) set_status(status string) {
	old_status := r.status
	r.status = status
	r.broadcast("status " + status)
	log.Printf("INFO race %s changed status %s -> %s", r.Race_code, old_status, status)
}

func (r *Race) run() {
	r.set_status("created")
	for {
		select {

		case <-r.countdown:
			log.Printf("INFO race %s started!", r.Race_code)

			r.lock.Lock()
			r.set_status("live")
			r.lock.Unlock()

		case msg := <-r.receive:
			r.lock.Lock()

			if msg.data[0] == 'p' {
				sender := r.players[msg.conn].player

				m := msg.data[2:]
				b := fmt.Sprintf("r %d %s", sender.player_id, m)
				log.Println("DEBUG broadcasting message: " + b)
				r.broadcast(b)
			}

			r.lock.Unlock()

		}
	}
}

func (r *Race) broadcast(message string) {
	for conn := range r.players {
		conn.send <- message
	}
}

func (r *Race) join(player *Player, ws *websocket.Conn) (*connection, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	conn := NewConnection(ws, player, r.receive)
	pp := &PlayerProgress{
		conn:   conn,
		player: player,
		race:   r,
		done:   0,
	}

	if len(r.players) == 0 {
	} else if r.start_time == nil {
		countdown_period := time.Second * 10
		start_time := time.Now().Add(countdown_period)
		r.start_time = &start_time

		log.Printf("INFO race %s set start time to %s", r.Race_code, r.start_time.Format("15:04:05.000"))
		go func() {
			<-time.After(start_time.Sub(time.Now()))
			r.countdown <- true
		}()
	} else {
		// do not allow any more player joins if there is
		// not enough time
		still_time_to_join := time.Second * 4
		if !time.Now().Before(
			r.start_time.Add(-still_time_to_join)) {
			return nil, errors.New("tried to join too late")
		}
	}

	r.players[pp.conn] = pp
	log.Printf("INFO player %s joined race %s", pp.player.name, r.Race_code)

	return conn, nil
}

func (r *Race) leave(player *Player) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, pp := range r.players {
		if pp.player == player {
			pp.conn.close()
			log.Printf("INFO player %s left race %s", player.name, r.Race_code)

			return nil
		}
	}

	return errors.New("player does not exist")
}
