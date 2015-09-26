package server

import (
	"bufio"
	"html/template"
	"io"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	letters        = "abcdefghijklmnopqrstuvwxyz0123456789"
	race_code_size = 7
	anonymous_name = "anonym"
)

var (
	templates *template.Template
)

type Lobby struct {
	texts []*string

	player_counter    int
	players           map[int]*Player
	races             map[string]*Race
	countdown_seconds int

	players_lock sync.Mutex
	races_lock   sync.Mutex
}

func NewLobby(tmpl *template.Template, texts_file string, countdown_seconds int) *Lobby {
	templates = tmpl

	f, err := os.Open(texts_file)
	defer f.Close()
	if err != nil {
		panic(err)
	}

	l := Lobby{
		players:           make(map[int]*Player),
		races:             make(map[string]*Race),
		texts:             make([]*string, 0),
		countdown_seconds: countdown_seconds,
	}

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		line = strings.Trim(line, "\n")
		l.texts = append(l.texts, &line)
	}

	return &l
}

func gen_code(size int) string {
	b := make([]byte, size)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b[:])
}

func (l *Lobby) create_race(practice bool) *Race {
	l.races_lock.Lock()
	defer l.races_lock.Unlock()

	var race_code string
	for {
		race_code = gen_code(race_code_size)
		if _, exists := l.races[race_code]; !exists {
			break
		}
	}

	race := NewRace(l, race_code, l.countdown_seconds, practice)
	text := *l.texts[rand.Intn(len(l.texts))]
	race.Race_string = &text
	race.race_text = []rune(text)

	l.races[race_code] = race
	go race.run()

	return race
}

func (l *Lobby) find_match_to_join() *Race {
	attempts := 10
	if len(l.races) < attempts {
		attempts = len(l.races)
	}

	i := 0
	for _, race := range l.races {
		if i >= attempts {
			break
		}

		if race.start_time == nil {
			return race
		}

		if time.Now().Before(race.start_time.Add(-3 * time.Second)) {
			return race
		}
	}

	return nil
}

func (l *Lobby) player_register() *Player {
	l.players_lock.Lock()
	defer l.players_lock.Unlock()

	l.player_counter++
	p := &Player{Player_id: l.player_counter}
	l.players[p.Player_id] = p
	return p
}

func (l *Lobby) player_sign_in(player_id int) *Player {
	l.players_lock.Lock()
	defer l.players_lock.Unlock()

	p, ok := l.players[player_id]
	if !ok {
		return nil
	}

	return p
}
