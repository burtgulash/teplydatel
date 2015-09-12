package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/pat"
)

var addr = flag.String("addr", ":1338", "http server address")
var templates = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()
	lobby := NewLobby("res/texts.txt")
	go lobby.run()

	r := pat.New()
	r.Get("/ws/{race_code}", lobby.ws_handler)
	r.Get("/zavod/{race_code}", lobby.race_handler)
	r.Get("/zavod", lobby.race_creator_handler)
	r.Get("/", lobby.lobby_handler)

	js := http.FileServer(http.Dir("js"))

	http.Handle("/js/", http.StripPrefix("/js/", js))
	http.Handle("/", r)

	log.Println("Serving on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
