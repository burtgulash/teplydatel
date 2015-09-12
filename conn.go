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
	send    chan []byte
	receive chan []byte
	alive   bool
}

func NewConnection(ws *websocket.Conn, receive chan<- []byte) *connection {
	return &connection{
		// 8 is size of buffer - # of messages before it gets full
		send:    make(chan []byte, 8),
		ws:      ws,
		receive: receive,
		alive:   true,
	}
}

func (conn *connection) run() {
	go conn.ws_reader()
	conn.ws_writer()
}

func (conn *connection) close() {
	alive = false
	close(conn.send)
}

func (conn *connection) ws_reader() {
	defer func() {
		conn.ws.Close()
	}()

	for alive {
		_, message, err := conn.ws.ReadMessage()
		if err != nil {
			break
		}
		conn.receive <- message
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

	for alive {
		message, ok := <-conn.send
		if !ok {
			conn.write(websocket.CloseMessage, []byte{})
			return
		}
		if err := conn.write(websocket.TextMessage, message); err != nil {
			return
		}
	}
}
