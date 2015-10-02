package server

import (
	"encoding/json"
	"errors"
	"time"
)

type icommand interface {
	get_conn() *connection
	set_conn(conn *connection)
}

type Command struct {
	conn *connection
}

type JSONCommand struct {
	Typ string `json:"typ"`
}

func (c *Command) get_conn() *connection {
	return c.conn
}

func (c *Command) set_conn(conn *connection) {
	c.conn = conn
}

func JSONDecode(data []byte, conn *connection) (icommand, error) {
	var x JSONUnion
	err := json.Unmarshal(data, &x)
	if err != nil {
		return nil, err
	}

	switch x.Typ {
	case "progress":
		return &x.ProgressCommand, nil
	case "start":
		return &x.StartCommand, nil
	}

	return nil, errors.New("couldn't parse command " + string(data))
}

type ProgressCommand struct {
	Command
	Done   string `json:"done"`
	Errors int    `json:"errors"`
}

type DisconnectCommand struct {
	Command
}

type StartCommand struct {
	Command
	Start_time time.Time `json:"at"`
}

type JSONUnion struct {
	JSONCommand
	ProgressCommand
	DisconnectCommand
	StartCommand
}
