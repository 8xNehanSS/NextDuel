package game_events

import (
	"encoding/json"
	"nextduel-svr/gamepkg"
	"nextduel-svr/server"
	"nextduel-svr/utils"
)

func HandleDuelAnswer(ctx *server.Context, player *gamepkg.Player, data json.RawMessage) {
	game := ctx.Game()
	room, exists := game.GetRoomByPlayerID(player.ID)
	if !exists {
		ctx.SendWSResponse([]*gamepkg.Player{player}, utils.GameActionError, "error: not in a room")
		return
	}

	hasAllAnswered, err := room.AnswerDuel(player.ID, data)
	if err != nil {
		ctx.SendWSResponse([]*gamepkg.Player{player}, utils.GameActionError, "error: "+err.Error())
		return
	}

	ctx.SendWSResponse([]*gamepkg.Player{player}, utils.GameActionDuelAnswer, "recieved")

	if !hasAllAnswered {
		ctx.SendWSResponse([]*gamepkg.Player{room.GetOpponent(player.ID)}, utils.GameActionDuelFast, "opponent answered")
	}
}
