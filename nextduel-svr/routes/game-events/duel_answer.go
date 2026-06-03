package game_events

import (
	"encoding/json"
	"nex-server/server"
	"nextduel-svr/gamepkg"
	"nextduel-svr/utils"
)

func HandleDuelAnswer(ctx *server.Context, player *gamepkg.Player, data json.RawMessage) {
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

	hasAllAnswered, err := room.AnswerDuel(player.ID, data)
	if err != nil {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "error: "+err.Error())
		return
	}

	ctx.SendWSResponse(player.Conn, utils.GameActionDuelAnswer, "recieved")

	if !hasAllAnswered {
		ctx.SendWSResponse(room.GetOpponent(player.ID).Conn, utils.GameActionDuelFast, "opponent answered")
	}
}
