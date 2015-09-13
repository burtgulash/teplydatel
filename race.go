package main

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Race struct {
	race_id      int64
	Race_code    string
	status       string
	created_time *time.Time
	start_time   *time.Time
	Race_text    *string
	lobby        *Lobby

	players      map[*Player]*connection
	players_lock sync.Mutex

	receive  chan []byte
	set_time chan time.Time
}

func NewRace(lobby *Lobby, race_code string) *Race {
	return &Race{
		lobby:     lobby,
		Race_code: race_code,
		players:   make(map[*Player]*connection),
		receive:   make(chan []byte),
		set_time:  make(chan time.Time),
	}
}

func (r *Race) set_status(status string) {
	old_status := r.status
	r.status = status
	r.broadcast("status " + status)
	log.Printf("race %s changed status %s -> %s", r.Race_code, old_status, status)
}

func (r *Race) run() {
	countdown := make(chan bool, 1)

	r.set_status("created")
	for {
		select {
		case <-countdown:
			r.set_status("live")
		case start_time := <-r.set_time:
			r.start_time = &start_time
			log.Printf("race %s set start time to %s", r.Race_code, r.start_time.Format("15:04:05.000"))
			go func() {
				<-time.After(start_time.Sub(time.Now()))
				countdown <- true
			}()
		}
	}
}

func (r *Race) broadcast(message string) {
	r.players_lock.Lock()
	defer r.players_lock.Unlock()

	for _, conn := range r.players {
		conn.send <- []byte(message)
	}
}

func (r *Race) join(player *Player, ws *websocket.Conn) (*connection, error) {
	conn := NewConnection(ws, r.receive)

	r.players_lock.Lock()
	defer r.players_lock.Unlock()

	r.players[player] = conn
	log.Printf("Player %s joined race %s", player.name, r.Race_code)

	if len(r.players) <= 1 {
	} else if r.start_time == nil {
		countdown_period := time.Second * 10
		r.set_time <- time.Now().Add(countdown_period)
	} else {
		still_time_to_join := time.Second * 4
		time.Now().Before(r.start_time.Add(-still_time_to_join))
	}

	return conn, nil
}

func (r *Race) leave(player *Player) error {
	r.players_lock.Lock()
	defer r.players_lock.Unlock()

	if conn, ok := r.players[player]; ok {
		delete(r.players, player)
		conn.close()

		log.Printf("Player %s left race %s", player.name, r.Race_code)
	}

	return nil
}
