package gamepkg

import (
	"nex-server/server"

	"github.com/gorilla/websocket"
)

func (gm *GameManager) HandleWSDisconnect(ctx *server.Context, conn *websocket.Conn, services any) error {
	playerID := ctx.UserID()
	player, exists := gm.GetPlayerByID(playerID)
	if !exists {
		return nil
	}

	gm.HandlePlayerDisconnect(player)

	return nil
}
