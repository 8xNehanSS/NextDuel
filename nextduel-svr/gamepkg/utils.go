package gamepkg

import (
	"nex-server/server"

	"github.com/gorilla/websocket"
)

func HandleWSDisconnect(ctx *server.Context, conn *websocket.Conn) error {
	game, ok := server.Get[*GameManager](ctx)
	if !ok {
		return nil
	}
	playerID := ctx.UserID()
	room, exists := game.GetRoomByPlayerID(playerID)
	if !exists {
		return nil // not in a room, nothing to clean up
	}

	room.HandleDisconnect(playerID)
	return nil
}
