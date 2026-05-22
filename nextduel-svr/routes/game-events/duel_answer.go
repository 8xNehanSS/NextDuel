package game_events

import (
	"encoding/json"
	"nextduel-svr/models"
	"nextduel-svr/server"
	"nextduel-svr/utils"
)

func HandleDuelAnswer(ctx *server.Context, player *models.Player, data json.RawMessage) {
	game := ctx.Game()
	room, exists := game.GetRoomByPlayerID(player.ID)
	if !exists {
		ctx.SendWSResponse([]*models.Player{player}, utils.GameActionError, "error: not in a room")
		return
	}

	hasAllAnswered, err := room.AnswerDuel(player.ID, data)
	if err != nil {
		ctx.SendWSResponse([]*models.Player{player}, utils.GameActionError, "error: "+err.Error())
		return
	}

	ctx.SendWSResponse([]*models.Player{player}, utils.GameActionDuelAnswer, "recieved")

	if !hasAllAnswered {
		ctx.SendWSResponse([]*models.Player{room.GetOpponent(player.ID)}, utils.GameActionDuelFast, "opponent answered")
	}
}
