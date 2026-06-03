package game_events

import (
	"nex-server/server"
	"nextduel-svr/gamepkg"
	"nextduel-svr/utils"
)

func HandleReadyPlayer(ctx *server.Context, player *gamepkg.Player, data any) {
	game, ok := server.Get[*gamepkg.GameManager](ctx)
	if !ok {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "game service not found")
		return
	}
	room, exists := game.GetRoomByPlayerID(player.ID)
	if !exists {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "error: not in a room")
		return
	}

	room.SetPlayerReady(player.ID)
	ctx.SendWSResponse(player.Conn, utils.GameActionPlayerReadyOk, "ok")
	room.TryStartGame()
}
