package main

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeTimeout = 5 * time.Second
)

type connection struct {
	ws      *websocket.Conn
	receive chan RaceMessage
	send    chan string
	alive   bool
}

func NewConnection(ws *websocket.Conn, player *Player, receive chan RaceMessage) *connection {
	return &connection{
		// 8 is size of buffer - # of messages before it gets full
		send:    make(chan string, 8),
		receive: receive,
		ws:      ws,
		alive:   true,
	}
}

func (conn *connection) run() {
	go conn.ws_reader()
	conn.ws_writer()
}

func (conn *connection) close() {
	conn.alive = false
	close(conn.send)
}

func (conn *connection) ws_reader() {
	defer func() {
		conn.ws.Close()
	}()

	for conn.alive {
		_, message, err := conn.ws.ReadMessage()
		if err != nil {
			break
		}

		conn.receive <- RaceMessage{
			conn: conn,
			data: string(message[:]),
		}
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

	for conn.alive {
		message, ok := <-conn.send
		if !ok {
			conn.write(websocket.CloseMessage, []byte{})
			return
		}
		if err := conn.write(websocket.TextMessage, []byte(message)); err != nil {
			return
		}
	}
}
