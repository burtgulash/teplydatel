package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	countdown_period = time.Second * 6
)

type RaceMessage struct {
	conn *connection
	data string
}

type progressItem struct {
	timestamp time.Time
	done      int
}

type PlayerProgress struct {
	conn   *connection
	player *Player
	race   *Race

	done        int
	doneHistory []*progressItem
	currentWpm  float64
}

func (pp *PlayerProgress) start(at time.Time) {
	pp.done = 0
	pp.doneHistory = []*progressItem{&progressItem{at, 0}}
	pp.currentWpm = 0.0
}

func (pp *PlayerProgress) add_progress(at time.Time, add int) {
	pp.done += add
	if pp.doneHistory == nil {
		panic("uninitialized doneHistory!")
	}
	first := pp.doneHistory[0]

	new_progress := &progressItem{at, pp.done}
	pp.doneHistory = append(pp.doneHistory, new_progress)
	pp.currentWpm = wpm(pp.done, time.Now().Sub(first.timestamp))
}

func wpm(num_characters int, period time.Duration) float64 {
	cps := float64(num_characters) / period.Seconds()
	wps := cps / 5.0
	return wps * 60.0
}

type Race struct {
	race_id      int64
	Race_code    string
	status       string
	created_time *time.Time
	start_time   *time.Time
	race_text    []rune
	Race_string  *string
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
	r.broadcast(fmt.Sprintf("s glob %s", status))
	log.Printf("INFO race %s changed status %s -> %s", r.Race_code, old_status, status)
}

func rune_equals(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (r *Race) run() {
	r.set_status("created")
	for {
		select {

		case <-r.countdown:
			log.Printf("INFO race %s started!", r.Race_code)

			r.lock.Lock()
			r.set_status("live")
			now := time.Now()
			for _, pp := range r.players {
				pp.start(now)
			}
			r.lock.Unlock()

		case msg := <-r.receive:
			r.lock.Lock()
			r.process(msg)
			r.lock.Unlock()

		}
	}
}

func (r *Race) broadcast(message string) {
	log.Println("DEBUG broadcasting message: " + message)
	for conn := range r.players {
		conn.send <- message
	}
}

func (r *Race) process(msg RaceMessage) {
	pp := r.players[msg.conn]

	if msg.data[0] == 'p' {
		m := []rune(msg.data[2:])
		text := r.race_text
		length := len(m)

		if rune_equals(text[pp.done:pp.done+length], m) {
			pp.add_progress(time.Now(), length)
		} else {
			// not matching, what do?
		}

		r.broadcast(fmt.Sprintf("r %d %d %.2f", pp.player.player_id, pp.done, pp.currentWpm))

		if pp.done == len(text) {
			r.broadcast(fmt.Sprintf("f %d", pp.player.player_id))
		}

	} else if msg.data == "disconnect" {
		pp.conn.close()
		delete(r.players, pp.conn)
		log.Printf("INFO player %s left race %s", pp.player.name, r.Race_code)
		r.broadcast(fmt.Sprintf("d %d", pp.player.player_id))

	}
}

func (r *Race) join(player *Player, ws *websocket.Conn) (*connection, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	conn := NewConnection(ws, player, r.receive)
	pp := &PlayerProgress{
		conn:        conn,
		player:      player,
		race:        r,
		done:        0,
		doneHistory: nil,
		currentWpm:  0.0,
	}

	if len(r.players) == 0 {
	} else if r.start_time == nil {
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
	r.broadcast(fmt.Sprintf("j %d %s", player.player_id, player.name))
	log.Printf("INFO player %s joined race %s", player.name, r.Race_code)

	return conn, nil
}
