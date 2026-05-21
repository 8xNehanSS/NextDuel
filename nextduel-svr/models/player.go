package models

import "github.com/gorilla/websocket"

type Player struct {
	ID    string
	Conn  *websocket.Conn
	Score int
}
