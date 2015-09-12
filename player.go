package main

type Player struct {
	player_id int
	name      string

	connections map[*string]*connection
}
