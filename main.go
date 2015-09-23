package main

import (
	"flag"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/pat"
	gcfg "gopkg.in/gcfg.v1"
)

var addr = flag.String("addr", ":1338", "http server address")
var config = flag.String("config", "config.ini", "config file")
var templates *template.Template

type Config struct {
	Texts struct {
		File string
	}
	Server struct {
		Address string
	}
	Website struct {
		Templates string
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	flag.Parse()

	var cfg Config
	err := gcfg.ReadFileInto(&cfg, *config)
	if err != nil {
		log.Fatalf("can't load config %s: %s", err)
	}

	addr := cfg.Server.Address
	templates, err = template.ParseGlob(cfg.Website.Templates + "/*")
	if err != nil {
		log.Fatalf("can't load templates: %s", err)
	}

	lobby := NewLobby(cfg.Texts.File)

	r := pat.New()
	r.Get("/ws/{race_code}", lobby.ws_handler)
	r.Get("/zavod/{race_code}", lobby.race_handler)
	r.Get("/zavod", lobby.race_creator_handler)
	r.Get("/", lobby.lobby_handler)

	js := http.FileServer(http.Dir("js"))

	http.Handle("/js/", http.StripPrefix("/js/", js))
	http.Handle("/", r)

	log.Println("INFO serving on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ERROR ListenAndServe: ", err)
	}
}
