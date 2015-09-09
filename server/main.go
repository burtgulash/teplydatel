package main

import (
	"github.com/bmizerany/pat"
	"log"
	"net/http"
)

func main() {
	lobby := NewLobby()
	go lobby.run()

	mux := pat.New()
	mux.Post("/ws", lobby.ws_handler)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
