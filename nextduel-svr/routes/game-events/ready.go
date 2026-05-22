package game_events

import (
	"nextduel-svr/models"
	"nextduel-svr/server"
)

func HandleReadyPlayer(ctx *server.Context, player *models.Player, data any) {
	game := ctx.Game()
	room, exists := game.GetRoomByPlayerID(player.ID)
	if !exists {
		return
	}

	room.SetPlayerReady(player.ID)

	if room.AreAllPlayersReady() {
		room.StartGame()
	}
	ctx.SendWSResponse([]*models.Player{player}, "player_ready", "ok")
}
