package game_events

import (
	"nextduel-svr/models"
	"nextduel-svr/server"
	"nextduel-svr/utils"
)

func HandleReadyPlayer(ctx *server.Context, player *models.Player, data any) {
	game := ctx.Game()
	room, exists := game.GetRoomByPlayerID(player.ID)
	if !exists {
		ctx.SendWSResponse([]*models.Player{player}, utils.GameActionError, "error: not in a room")
		return
	}

	room.SetPlayerReady(player.ID)
	ctx.SendWSResponse([]*models.Player{player}, utils.GameActionPlayerReady, "ok")
	room.TryStartGame()
}
