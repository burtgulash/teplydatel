package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/pat"
	gcfg "gopkg.in/gcfg.v1"

	"teplydatel/server"
)

var addr = flag.String("addr", ":1338", "http server address")
var config = flag.String("config", "config.ini", "config file")

type Config struct {
	Texts struct {
		File string
	}
	Server struct {
		Address   string
		Templates string
		Static    []string
	}
	Race struct {
		CountdownSeconds int
	}
}

func parse_static_location(loc_string string) (string, string, error) {
	sp := strings.Split(loc_string, "->")
	if len(sp) != 2 {
		return "", "", fmt.Errorf("correct format = {endpoint} -> {path}. Got: %s", loc_string)
	}

	endpoint := strings.Trim(sp[0], " /")
	dir := strings.Trim(sp[1], " ")

	return "/" + endpoint + "/", dir, nil
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	flag.Parse()

	var cfg Config
	err := gcfg.ReadFileInto(&cfg, *config)
	if err != nil {
		log.Fatalf("ERROR can't load config %s: %s", err)
	}

	addr := cfg.Server.Address
	templates, err := template.ParseGlob(cfg.Server.Templates + "/*")
	if err != nil {
		log.Fatalf("ERROR can't load templates: %s", err)
	}

	lobby := server.NewLobby(templates, cfg.Texts.File, cfg.Race.CountdownSeconds)

	r := pat.New()
	r.Get("/ws/{race_code}", lobby.Ws_handler)
	r.Get("/zavod/{race_code}", lobby.Race_handler)
	r.Get("/zavod", lobby.Race_creator_handler)
	r.Get("/", lobby.Lobby_handler)

	// get all static directories from config and create fileserver
	// for each one of them
	for _, s := range cfg.Server.Static {
		endpoint, dir, err := parse_static_location(s)
		if err != nil {
			log.Fatalf("ERROR %s", err)
		}

		fileserver := http.FileServer(http.Dir(dir))
		http.Handle(endpoint, http.StripPrefix(endpoint, fileserver))
		log.Printf("INFO serving static files from %s on %s", dir, endpoint)
	}

	http.Handle("/", r)

	log.Println("INFO serving on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ERROR ListenAndServe: ", err)
	}
}
