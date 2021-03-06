package server

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeTimeout   = 20 * time.Second
	pongWait       = 20 * time.Second
	pingPeriod     = pongWait * 4 / 5
	maxMessageSize = 64
)

type connection struct {
	ws      *websocket.Conn
	receive chan icommand
	send    chan []byte
	alive   bool
}

func NewConnection(ws *websocket.Conn, player *Player, receive chan icommand) *connection {
	return &connection{
		// 8 is size of buffer - # of messages before it gets full
		send:    make(chan []byte, 8),
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
		conn.receive <- &DisconnectCommand{Command{conn}}
		conn.ws.Close()
	}()

	conn.ws.SetReadLimit(maxMessageSize)
	conn.ws.SetReadDeadline(time.Now().Add(pongWait))
	conn.ws.SetPongHandler(func(string) error {
		conn.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for conn.alive {
		_, message, err := conn.ws.ReadMessage()
		if err != nil {
			break
		}

		cmd_message, err := JSONDecode(message, conn)
		if err != nil {
			log.Printf("ERROR couldn't parse json message: %v", err)
			break
		}

		cmd_message.set_conn(conn)
		conn.receive <- cmd_message
	}
}

func (conn *connection) write(mt int, message []byte) error {
	conn.ws.SetWriteDeadline(time.Now().Add(writeTimeout))
	return conn.ws.WriteMessage(mt, message)
}

func (conn *connection) ws_writer() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.ws.Close()
	}()

	for conn.alive {
		select {
		case message, ok := <-conn.send:
			if !ok {
				conn.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.write(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			if err := conn.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
