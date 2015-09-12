package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/drone/routes"
)

var addr = flag.String("addr", ":1338", "http server address")
var templates = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	flag.Parse()
	lobby := NewLobby()
	go lobby.run()

	mux := routes.New()
	mux.Get("/ws/:race_code", lobby.ws_handler)
	mux.Get("/zavod/:race_code", lobby.race_handler)
	mux.Get("/zavod", lobby.race_creator_handler)
	mux.Get("/", lobby.lobby_handler)

	pwd, _ := os.Getwd()
	mux.Static("/js", pwd)

	http.Handle("/", mux)

	log.Println("Serving on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
