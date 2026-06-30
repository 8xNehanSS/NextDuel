package game_events

import (
	"nex-server/server"
	"nextduel-svr/gamepkg"
	"nextduel-svr/utils"
)

func (g *GameEventServices) HandleReadyPlayer(ctx *server.Context, player *gamepkg.Player, data any) {
	if g.Game == nil {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "game service not found")
		return
	}
	room, exists := g.Game.GetRoomByPlayerID(player.ID)
	if !exists {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "error: not in a room")
		return
	}

	room.SetPlayerReady(player.ID)
	ctx.SendWSResponse(player.Conn, utils.GameActionPlayerReadyOk, "ok")
	room.TryStartGame()
}
