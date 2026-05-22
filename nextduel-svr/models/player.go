package models

import "github.com/gorilla/websocket"

type Player struct {
	ID    string
	Conn  *websocket.Conn
	Score int
	Ready bool
}

func (p *Player) NewPlayer(id string, conn *websocket.Conn) *Player {
	return &Player{
		ID:    id,
		Conn:  conn,
		Score: 0,
		Ready: false,
	}
}

func (p *Player) SetReady(ready bool) {
	p.Ready = ready
}
