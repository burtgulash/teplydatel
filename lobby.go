package main

import (
	"bufio"
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
)

type Lobby struct {
	texts []*string

	players map[*Player]bool
	races   map[string]*Race

	players_lock sync.Mutex
	races_lock   sync.Mutex
}

func NewLobby(texts_file string) *Lobby {
	f, err := os.Open(texts_file)
	defer f.Close()
	if err != nil {
		panic(err)
	}

	l := Lobby{
		players: make(map[*Player]bool),
		races:   make(map[string]*Race),
		texts:   make([]*string, 0),
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

func (l *Lobby) create_race() *Race {
	l.races_lock.Lock()
	defer l.races_lock.Unlock()

	var race_code string
	for {
		race_code = gen_code(race_code_size)
		if _, exists := l.races[race_code]; !exists {
			break
		}
	}

	race := NewRace(l, race_code)
	text := l.texts[rand.Intn(len(l.texts))]
	race.Race_text = text

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

		if time.Now().Before(race.start_time.Add(-5 * time.Second)) {
			return race
		}
	}

	return nil
}
