package gamepkg

import "github.com/gorilla/websocket"

type Player struct {
	ID    string
	Conn  *websocket.Conn
	score int
	ready bool

	answer any
}

func (p *Player) NewPlayer(id string, conn *websocket.Conn) *Player {
	return &Player{
		ID:    id,
		Conn:  conn,
		score: 0,
		ready: false,
	}
}

func (p *Player) GetScore() int {
	return p.score
}

func (p *Player) IncrementScore(score int) {
	p.score += score
}

func (p *Player) DecrementScore(score int) {
	p.score -= score
}

func (p *Player) IsReady() bool {
	return p.ready
}

func (p *Player) SetReady(ready bool) {
	p.ready = ready
}

func (p *Player) SetAnswer(answer any) {
	p.answer = answer
}

func (p *Player) GetAnswer() any {
	return p.answer
}

func (p *Player) ClearAnswer() {
	p.answer = nil
}
