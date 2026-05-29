package game_events

import (
	"nextduel-svr/gamepkg"
	"nextduel-svr/server"
	"nextduel-svr/utils"
)

func HandleReadyPlayer(ctx *server.Context, player *gamepkg.Player, data any) {
	game := ctx.Game()
	room, exists := game.GetRoomByPlayerID(player.ID)
	if !exists {
		ctx.SendWSResponse([]*gamepkg.Player{player}, utils.GameActionError, "error: not in a room")
		return
	}

	room.SetPlayerReady(player.ID)
	ctx.SendWSResponse([]*gamepkg.Player{player}, utils.GameActionPlayerReady, "ok")
	room.TryStartGame()
}
