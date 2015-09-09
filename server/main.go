package main

import (
	"flag"
	"github.com/drone/routes"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":1338", "http server address")

func main() {
	flag.Parse()
	lobby := NewLobby()
	go lobby.run()

	mux := routes.New()
	mux.Post("/ws/:race_code([a-z0-9]{7})", lobby.ws_handler)

	log.Println("Serving on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
