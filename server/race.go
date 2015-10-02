package server

import (
	"errors"
	"log"
	"regexp"
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
	color    string

	done       int
	errors     int
	history    []*progressItem
	currentWpm float64
	lastWpm    float64
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

	first := pp.history[0]
	now := time.Now()

	pp.currentWpm = wpm(pp.done, now.Sub(first.timestamp))

	sum := 0.0
	c := 0
	for i := len(pp.history) - 1; i > 0 && c < 7; i-- {
		a, b := pp.history[i], pp.history[i-1]
		x := wpm(a.num_ok, a.timestamp.Sub(b.timestamp))
		sum += x
		c++
	}

	if c > 0 {
		log.Println(sum, c)
		pp.lastWpm = sum / float64(c)
	} else {
		pp.lastWpm = 0
	}

	new_progress := &progressItem{at, num_ok, num_errs}
	pp.history = append(pp.history, new_progress)
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

	receive   chan icommand
	countdown chan int
	start_it  chan bool
}

func NewRace(lobby *Lobby, race_code string, countdown_seconds int, practice bool) *Race {
	return &Race{
		lobby:            lobby,
		Race_code:        race_code,
		players:          make(map[*connection]*PlayerProgress),
		receive:          make(chan icommand, 16),
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

	r.broadcast(msg_status(0, new_status))
	log.Printf("INFO race changed status {race=%s, from=%s, to=%s}", r.Race_code, old_status, new_status)
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
			log.Printf("closing race {race=%s}", r.Race_code)
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
			r.broadcast(msg_countdown(0, remains))
			r.lock.Unlock()

		case cmd := <-r.receive:
			r.lock.Lock()
			pp, ok := r.players[cmd.get_conn()]
			if !ok {
				log.Println("ERROR player progress not found {conn=%v}", cmd.get_conn())
				r.lock.Unlock()
				break
			}

			switch cmd := cmd.(type) {
			case *ProgressCommand:
				r.handle_progress(pp, cmd.Done, cmd.Errors)
			case *StartCommand:
				if r.is_practice_race {
					r.start_it <- true
				} else {
					log.Printf("non-pratice race attempted to be started as practice race {race=%s}", r.Race_code)
				}
			case *DisconnectCommand:
				pp.conn.close()
				delete(r.players, pp.conn)

				log.Printf("INFO player left race {player=%d, race=%s}", pp.player.Player_id, r.Race_code)
				r.broadcast(msg_disconnected(pp.player.Player_id))

				if len(r.players) == 0 {
					r.set_status(CLOSED)
				}
			}
			r.lock.Unlock()

		}
	}
}

func (r *Race) broadcast(message []byte) {
	log.Println("DEBUG broadcasting: " + string(message))
	for conn := range r.players {
		conn.send <- message
	}
}

func (r *Race) handle_progress(pp *PlayerProgress, done string, num_errors int) {
	if pp.finished {
		log.Printf("WARNING player already finished, but progress received {player=%d, race=%s}", pp.player.Player_id, r.Race_code)
		return
	}

	m := []rune(done[:])
	r.progress(pp, num_errors, m)

	if pp.done == len(r.race_text) {
		if !pp.finished {
			pp.finished = true
			r.handle_finished(pp)
		} else {
			log.Printf("WARNING player finished more than once! {player=%d}", pp.player.Player_id)
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

	r.broadcast(msg_progress(pp.player.Player_id, pp.done, pp.errors, pp.lastWpm))
}

func (r *Race) handle_finished(pp *PlayerProgress) {
	pp.rank = r.next_rank
	r.broadcast(msg_player_finished(pp.player.Player_id, pp.rank))
	r.next_rank++
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

	num_players := len(r.players)
	color := player_color_palette[num_players%len(player_color_palette)]

	conn := NewConnection(ws, player, r.receive)
	pp := &PlayerProgress{
		conn:       conn,
		player:     player,
		race:       r,
		done:       0,
		history:    nil,
		currentWpm: 0.0,
		color:      color,
	}

	if r.is_practice_race {
	} else if len(r.players) == 0 {
		// wait for other players to join
	} else if r.start_time == nil {
		r.start_countdown(r.countdown_period)
		log.Printf("INFO race set start time {race=%s, start_time=%s}", r.Race_code, r.start_time.Format("15:04:05.000"))
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
		conn.send <- msg_player_joined(pp.player.Player_id, pp.color)
	}

	r.players[pp.conn] = pp
	r.broadcast(msg_player_joined(pp.player.Player_id, pp.color))
	log.Printf("INFO player joined race {player=%d, race=%s}", player.Player_id, r.Race_code)

	return conn, nil
}
