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
	created_time time.Time
	start_time   time.Time
	Race_text    *string
	lobby        *Lobby

	players      map[*Player]*connection
	players_lock *sync.Mutex

	receive chan []byte
}

func NewRace(lobby *Lobby, race_code string) *Race {
	return &Race{
		lobby:     lobby,
		Race_code: race_code,
		players:   make(map[*Player]bool),
		receive:   make(chan []byte),
	}
}

func (r *Race) run() {
	for event := range r.receive {
		// do shit with incoming messages
		// broadcast to all participants
	}
}

func (r *Race) broadcast(message []byte) {
	r.players_lock.Lock()
	defer r.players_lock.Unlock()

	for _, conn := range r.players {
		conn.send <- message
	}
}

func (r *Race) join(player *Player, ws *websocket.Conn) error {
	conn := NewConnection(ws, r.receive)

	r.players_lock.Lock()
	defer r.players_lock.Unlock()

	r.players[player] = conn
	log.Printf("Player %s joined race %s", player.name, r.Race_code)

	go conn.run()

	return nil
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
