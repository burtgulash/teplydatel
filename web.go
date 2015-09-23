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
	blockKey      = []byte("migakus-sakretus")
	secure_cooker = securecookie.New(hashKey, blockKey)

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func get_cookie(r *http.Request, cookie_name string) *http.Cookie {
	cookie, err := r.Cookie(cookie_name)
	if err != nil {
		return nil
	}
	return cookie
}

func decode_cookie(cookie *http.Cookie, cookie_name string) map[string]int {
	value := make(map[string]int)
	err := secure_cooker.Decode(cookie_name, cookie.Value, &value)
	if err != nil {
		return nil
	}

	return value
}

func set_auth_cookie(w http.ResponseWriter, player *Player) error {
	value := map[string]int{
		"player_id": player.Player_id,
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
	templates.ExecuteTemplate(w, "index.html", nil)
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

	var player *Player
	new_registration := false
	cookie := get_cookie(r, "auth")
	if cookie == nil {
		new_registration = true
	} else {
		cookie_value := decode_cookie(cookie, "auth")
		if cookie_value == nil {
			http.Error(w, "Race not found", http.StatusInternalServerError)
			return
		}

		player = l.player_sign_in(cookie_value["player_id"])
		if player == nil {
			new_registration = true
		}
	}
	if new_registration {
		player = l.player_register()
		err := set_auth_cookie(w, player)
		if err != nil {
			http.Error(w, "server error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	templates.ExecuteTemplate(w, "race.html", struct {
		Race   *Race
		Player *Player
	}{
		race,
		player,
	})
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

	cookie := get_cookie(r, "auth")
	if cookie == nil {
		ws.WriteMessage(websocket.CloseMessage, []byte("Unauthenticated"))
		ws.Close()
		return
	}

	cookie_value := decode_cookie(cookie, "auth")
	if cookie_value == nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	player := l.player_sign_in(cookie_value["player_id"])
	if player == nil {
		ws.WriteMessage(websocket.CloseMessage, []byte("Forbidden"))
		ws.Close()
		return
	}

	conn, err := race.join(player, ws)
	if err != nil {
		ws.WriteMessage(websocket.CloseMessage, []byte("Forbidden to join the race"))
		ws.Close()
		return
	}

	conn.run()
}
