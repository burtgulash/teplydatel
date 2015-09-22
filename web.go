package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/websocket"
)

var (
	hashKey       = []byte("sekretus-magikus")
	blockKey      = []byte("migakus-saekretus")
	secure_cooker = securecookie.New(hashKey, blockKey)

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func set_auth_cookie(w http.ResponseWriter, player *Player) error {
	value := map[string]int{
		"player_id": player.player_id,
	}
	encoded, err := secure_cooker.Encode("auth", value)
	if err != nil {
		return errors.New("can't encode cookie: " + err.Error())
	}
	cookie := &http.Cookie{
		Name:  "auth",
		Value: encoded,
		Path:  "/",
	}
	http.SetCookie(w, cookie)

	return nil
}

func (l *Lobby) lobby_handler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", struct{ Name string }{"John"})
}

func (l *Lobby) race_creator_handler(w http.ResponseWriter, r *http.Request) {
	race_type := r.URL.Query().Get("t")

	var race *Race
	if race_type == "verejny" {
		race = l.find_match_to_join()
		if race == nil {
			race = l.create_race()
		}
	} else if race_type == "soukromy" {
		race = l.create_race()
	} else {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/zavod/"+race.Race_code, http.StatusFound)
}

func (l *Lobby) race_handler(w http.ResponseWriter, r *http.Request) {
	race_code := r.URL.Query().Get(":race_code")
	race, exists := l.races[race_code]
	if !exists {
		http.Error(w, "Race not found", http.StatusNotFound)
		return
	}

	templates.ExecuteTemplate(w, "race.html", race)
}

func (l *Lobby) ws_handler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	race_code := r.URL.Query().Get(":race_code")

	race, in := l.races[race_code]
	if !in {
		ws.WriteMessage(websocket.CloseMessage, []byte("Race does not exist"))
		ws.Close()
		return
	}

	var player *Player
	cookie, err := r.Cookie("auth")
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	value := make(map[string]int)
	err = secure_cooker.Decode("auth", cookie.Value, &value)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	player = l.player_sign_in(value["player_id"])
	if player == nil {
		player = l.player_register()
		err := set_auth_cookie(w, player)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
	}

	log.Println("PLAYEA", player)

	conn, err := race.join(player, ws)
	if err != nil {
		ws.WriteMessage(websocket.CloseMessage, []byte("Forbidden to join the race"))
		ws.Close()
		return
	}

	conn.run()
}
