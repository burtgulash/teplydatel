package main

import (
	"flag"
	"fmt"
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
	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello wordl!")
	})
	mux.Post("/ws/:race_code", lobby.ws_handler)

	http.Handle("/", mux)

	log.Println("Serving on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
