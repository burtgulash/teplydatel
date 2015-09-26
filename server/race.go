package server

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	CREATED = iota
	LIVE
	CLOSED
)

var (
	progress_rx = regexp.MustCompile("(\\d+) (.*)")
	status_map  = map[int]string{
		CREATED: "created",
		LIVE:    "live",
		CLOSED:  "closed",
	}
)

type RaceMessage struct {
	conn *connection
	data string
}

type progressItem struct {
	timestamp time.Time
	num_ok    int
	num_errs  int
}

type PlayerProgress struct {
	conn     *connection
	player   *Player
	race     *Race
	rank     int
	finished bool

	done       int
	errors     int
	history    []*progressItem
	currentWpm float64
}

func (pp *PlayerProgress) start(at time.Time) {
	pp.done = 0
	pp.history = []*progressItem{&progressItem{at, 0, 0}}
	pp.currentWpm = 0.0
}

func (pp *PlayerProgress) add_progress(at time.Time, num_ok, num_errs int) {
	pp.done += num_ok
	pp.errors += num_errs

	if pp.history == nil {
		panic("uninitialized history!")
	}

	new_progress := &progressItem{at, num_ok, num_errs}
	pp.history = append(pp.history, new_progress)

	first := pp.history[0]
	pp.currentWpm = wpm(pp.done, time.Now().Sub(first.timestamp))
}

func wpm(num_characters int, period time.Duration) float64 {
	cps := float64(num_characters) / period.Seconds()
	wps := cps / 5.0
	return wps * 60.0
}

type Race struct {
	race_id          int64
	Race_code        string
	status           int
	created_time     *time.Time
	start_time       *time.Time
	race_text        []rune
	Race_string      *string
	lobby            *Lobby
	next_rank        int
	countdown_period time.Duration
	is_practice_race bool

	players map[*connection]*PlayerProgress
	lock    sync.Mutex

	receive   chan RaceMessage
	countdown chan int
	start_it  chan bool
}

func NewRace(lobby *Lobby, race_code string, countdown_seconds int, practice bool) *Race {
	return &Race{
		lobby:            lobby,
		Race_code:        race_code,
		players:          make(map[*connection]*PlayerProgress),
		receive:          make(chan RaceMessage, 16),
		countdown:        make(chan int),
		start_it:         make(chan bool, 1),
		next_rank:        1,
		countdown_period: time.Second * time.Duration(countdown_seconds),
		is_practice_race: practice,
	}
}

func (r *Race) set_status(status int) {
	old_status := status_map[r.status]
	new_status := status_map[status]

	r.status = status

	r.broadcast(fmt.Sprintf("s glob %s", new_status))
	log.Printf("INFO race changed status %s -> %s {race=%s}", old_status, new_status, r.Race_code)
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
	r.set_status(CREATED)

	for {
		if r.status == CLOSED {
			log.Printf("closing race. {race=%s}", r.Race_code)
			break
		}

		select {

		case <-r.start_it:
			r.lock.Lock()
			r.set_status(LIVE)
			now := time.Now()
			for _, pp := range r.players {
				pp.start(now)
			}
			r.lock.Unlock()

		case remains := <-r.countdown:
			r.lock.Lock()
			r.broadcast(fmt.Sprintf("c glob %d", remains))
			r.lock.Unlock()

		case msg := <-r.receive:
			r.lock.Lock()
			pp := r.players[msg.conn]

			if msg.data[0] == 'p' {
				r.handle_progress(pp, msg.data)

			} else if msg.data == "disconnect" {
				pp.conn.close()
				delete(r.players, pp.conn)

				log.Printf("INFO player left race {player=%d, race=%s}", pp.player.Player_id, r.Race_code)
				r.broadcast(fmt.Sprintf("d %d", pp.player.Player_id))

				if len(r.players) == 0 {
					r.set_status(CLOSED)
				}
			}
			r.lock.Unlock()

		}
	}
}

func (r *Race) broadcast(message string) {
	log.Println("DEBUG broadcasting: " + message)
	for conn := range r.players {
		conn.send <- message
	}
}

func (r *Race) handle_progress(pp *PlayerProgress, msg string) {
	if pp.finished {
		log.Printf("WARNING player already finished, but progress received. {player=%d, race=%s}", pp.player.Player_id, r.Race_code)
		return
	}
	m := progress_rx.FindStringSubmatch(msg[2:])

	num_errors, err := strconv.Atoi(m[1])
	if err != nil {
		return
	}

	r.progress(pp, num_errors, []rune(m[2]))

	if pp.done == len(r.race_text) {
		if !pp.finished {
			pp.finished = true
			r.handle_finished(pp)
		} else {
			log.Printf("WARNING player finished more than once!. {player=%d}", pp.player.Player_id)
		}
	}
}

func (r *Race) progress(pp *PlayerProgress, num_errors int, msg []rune) {
	text := r.race_text
	length := len(msg)

	if rune_equals(text[pp.done:pp.done+length], msg) {
		pp.add_progress(time.Now(), length, num_errors)
	} else {
		// not matching, what do?
	}

	r.broadcast(fmt.Sprintf("r %d %d %d %.2f", pp.player.Player_id, pp.done, pp.errors, pp.currentWpm))
}

func (r *Race) handle_finished(pp *PlayerProgress) {
	pp.rank = r.next_rank
	r.broadcast(fmt.Sprintf("f %d %d", pp.player.Player_id, pp.rank))
	r.next_rank++
}

func notification_player_joined(player *Player) string {
	return fmt.Sprintf("j %d %d", player.Player_id, player.Player_id)
}

func (r *Race) start_countdown(countdown_period time.Duration) {
	start_time := time.Now().Add(r.countdown_period)
	r.start_time = &start_time

	go func() {
		to_start := r.start_time.Sub(time.Now()) / time.Millisecond
		remaining := int(to_start) / 1000
		round := int(to_start) % 1000
		<-time.After(time.Millisecond * time.Duration(round))

		for i := remaining; i > 0; i-- {
			r.countdown <- i
			time.Sleep(time.Second)
		}
	}()

	go func() {
		<-time.After(start_time.Sub(time.Now()))
		r.start_it <- true
	}()
}

func (r *Race) join(player *Player, ws *websocket.Conn) (*connection, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	conn := NewConnection(ws, player, r.receive)
	pp := &PlayerProgress{
		conn:       conn,
		player:     player,
		race:       r,
		done:       0,
		history:    nil,
		currentWpm: 0.0,
	}

	if r.is_practice_race {
		// TODO change countdown period for practice races
		r.start_countdown(r.countdown_period)
		log.Printf("INFO practice race set start time to %s. {race=%s}", r.start_time.Format("15:04:05.000"), r.Race_code)
	} else if len(r.players) == 0 {
		// wait for other players to join
	} else if r.start_time == nil {
		r.start_countdown(r.countdown_period)
		log.Printf("INFO race set start time to %s. {race=%s}", r.start_time.Format("15:04:05.000"), r.Race_code)
	} else {
		// do not allow any more player joins if there is
		// not enough time
		still_time_to_join := time.Second * 4
		if !time.Now().Before(
			r.start_time.Add(-still_time_to_join)) {
			return nil, errors.New("tried to join too late")
		}
	}

	// notify current user of all joined users
	for _, pp := range r.players {
		conn.send <- notification_player_joined(pp.player)
	}

	r.players[pp.conn] = pp
	// TODO remove 2x Player_id
	r.broadcast(notification_player_joined(player))
	log.Printf("INFO player joined race. {player=%d, race=%s}", player.Player_id, r.Race_code)

	return conn, nil
}

func (r *Race) join_practice_race(player *Player, ws *websocket.Conn) {
	r.lock.Lock()
	defer r.lock.Unlock()
}
