package server

import "encoding/json"

const ROOT_PLAYER_ID = 0

type Message struct {
	Type      string `json:"typ"`
	Player_id int    `json:"plid"`
}

type JoinMessage struct {
	Message
	Color string `json:"color"`
}

func msg_player_joined(player_id int, color string) []byte {
	x, _ := json.Marshal(JoinMessage{
		Message{
			"joined",
			player_id,
		},
		color,
	})
	return x
}

type FinishedMessage struct {
	Message
	Rank int `json:"rank"`
}

func msg_player_finished(player_id, rank int) []byte {
	x, _ := json.Marshal(FinishedMessage{
		Message{
			"finished",
			player_id,
		},
		rank,
	})
	return x
}

type ProgressMessage struct {
	Message
	Done   int     `json:"done"`
	Errors int     `json:"errors"`
	Wpm    float64 `json:"wpm"`
}

func msg_progress(player_id, done, errors int, wpm float64) []byte {
	x, _ := json.Marshal(ProgressMessage{
		Message{
			"progress",
			player_id,
		},
		done,
		errors,
		RoundN(wpm, 2),
	})
	return x
}

type DisconnectedMessage struct {
	Message
}

func msg_disconnected(player_id int) []byte {
	x, _ := json.Marshal(DisconnectedMessage{
		Message{
			"disconnected",
			player_id,
		},
	})
	return x
}

type CountdownMessage struct {
	Message
	Remains int `json:"remains"`
}

func msg_countdown(player_id, remains int) []byte {
	x, _ := json.Marshal(CountdownMessage{
		Message{
			"countdown",
			player_id,
		},
		remains,
	})
	return x
}

type StatusMessage struct {
	Message
	Status string `json:"status"`
}

func msg_status(player_id int, status string) []byte {
	x, _ := json.Marshal(StatusMessage{
		Message{
			"status",
			player_id,
		},
		status,
	})
	return x
}

type RaceInfoMessage struct {
	Message
	Code      string `json:"code"`
	Race_type string `json:"race_type"`
}

func msg_race_info(race_code string, race_type string) []byte {
	x, _ := json.Marshal(RaceInfoMessage{
		Message{
			"info",
			ROOT_PLAYER_ID,
		},
		race_code,
		race_type,
	})
	return x
}
