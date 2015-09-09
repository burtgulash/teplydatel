package main

import (
	"github.com/gorilla/websocket"
	"time"
)

const (
	writeTimeout = 5 * time.Second
)

type connection struct {
	ws   *websocket.Conn
	send chan []byte
	race *Race
}

func (conn *connection) reader() {
	defer func() {
		conn.race.unregister <- conn
		conn.ws.Close()
	}()

	for {
		if _, message, err := conn.ws.ReadMessage(); err != nil {
			break
		}
		conn.race.receive <- message
	}
}

func (conn *connection) write(mt int, message []byte) error {
	conn.ws.SetWriteDeadline(time.Now().Add(writeTimeout))
	return conn.ws.WriteMessage(mt, message)
}

func (conn *connection) writer() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.ws.Close()
	}()

	for {
		select {
		case message, ok := conn.send:
			if !ok {
				conn.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err != conn.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
