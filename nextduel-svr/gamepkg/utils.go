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
	player, exists := game.GetPlayerByID(playerID)
	if !exists {
		return nil
	}

	game.HandlePlayerDisconnect(player)

	return nil
}
