package main

import (
	"bufio"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
)

const (
	letters        = "abcdefghijklmnopqrstuvwxyz0123456789"
	race_code_size = 7
)

type Lobby struct {
	players map[*Player]bool
	races   map[string]*Race
	texts   []*string

	create_race     chan chan *Race
	unregister_race chan *Race
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

		create_race:     make(chan chan *Race),
		unregister_race: make(chan *Race),
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

func (l *Lobby) run() {
	log.Println("Lobby running")
	for {
		select {
		case race_request_req := <-l.create_race:
			race := l.make_race()
			l.races[race.Race_code] = race

			// Run the race!
			go race.run()
			log.Println("Created race", race.Race_code)

			// Send the newly created race back to requester
			race_request_req <- race
			close(race_request_req)
		case race := <-l.unregister_race:
			if _, in := l.races[race.Race_code]; in {
				delete(l.races, race.Race_code)
			} else {
				log.Println("ERROR", "can't unregister non-existing race", race.Race_code)
			}
		}
	}
}

func gen_code(size int) string {
	b := make([]byte, size)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b[:])
}

func (l *Lobby) make_race() *Race {
	// TODO create race_code here
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

	return race
}

func (l *Lobby) Create_private_race() *Race {
	req := make(chan *Race, 1)
	l.create_race <- req
	race := <-req
	return race
}

func (l *Lobby) Create_or_join_public_race() *Race {
	return l.Create_private_race()
}
