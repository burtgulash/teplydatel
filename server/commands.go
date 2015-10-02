package server

import "encoding/json"

type Command struct {
	Cmd       string `json:"cmd"`
	Player_id int    `json:"plid"`
}

type JoinMessage struct {
	Command
	Color string `json:"color"`
}

func cmd_player_joined(player_id int, color string) []byte {
	x, _ := json.Marshal(JoinMessage{
		Command{
			"joined",
			player_id,
		},
		color,
	})
	return x
}

type FinishedMessage struct {
	Command
	Rank int `json:"rank"`
}

func cmd_player_finished(player_id, rank int) []byte {
	x, _ := json.Marshal(FinishedMessage{
		Command{
			"finished",
			player_id,
		},
		rank,
	})
	return x
}

type ProgressMessage struct {
	Command
	Done   int     `json:"done"`
	Errors int     `json:"errors"`
	Wpm    float64 `json:"wpm"`
}

func cmd_progress(player_id, done, errors int, wpm float64) []byte {
	x, _ := json.Marshal(ProgressMessage{
		Command{
			"progress",
			player_id,
		},
		done,
		errors,
		wpm,
	})
	return x
}

type DisconnectedMessage struct {
	Command
}

func cmd_disconnected(player_id int) []byte {
	x, _ := json.Marshal(DisconnectedMessage{
		Command{
			"disconnected",
			player_id,
		},
	})
	return x
}

type CountdownMessage struct {
	Command
	Remains int `json:"remains"`
}

func cmd_countdown(player_id, remains int) []byte {
	x, _ := json.Marshal(CountdownMessage{
		Command{
			"countdown",
			player_id,
		},
		remains,
	})
	return x
}

type StatusMessage struct {
	Command
	Status string `json:"status"`
}

func cmd_status(player_id int, status string) []byte {
	x, _ := json.Marshal(StatusMessage{
		Command{
			"status",
			player_id,
		},
		status,
	})
	return x
}
