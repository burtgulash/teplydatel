package main

import (
	"github.com/gorilla/websocket"
	"time"
)

const (
	writeTimeout = 5 * time.Second
)

type connection struct {
	ws     *websocket.Conn
	send   chan []byte
	player *Player
	race   *Race
}

func (conn *connection) ws_reader() {
	defer func() {
		conn.race.unregister <- conn.player
		conn.ws.Close()
	}()

	for {
		_, message, err := conn.ws.ReadMessage()
		if err != nil {
			break
		}
		conn.race.receive <- message
	}
}

func (conn *connection) write(mt int, message []byte) error {
	conn.ws.SetWriteDeadline(time.Now().Add(writeTimeout))
	return conn.ws.WriteMessage(mt, message)
}

func (conn *connection) ws_writer() {
	defer func() {
		conn.ws.Close()
	}()

	for {
		select {
		case message, ok := <-conn.send:
			if !ok {
				conn.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.write(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}
