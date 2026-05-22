package game_events

import (
	"nextduel-svr/models"
	"nextduel-svr/server"
	"nextduel-svr/utils"
)

func HandleJoinQueue(ctx *server.Context, player *models.Player, data any) {
	game := ctx.Game()
	room, matched, err := game.AddToQueue(player)
	if err != nil {
		ctx.SendWSResponse([]*models.Player{player}, utils.GameActionError, err.Error())
		return
	}
	if matched {
		ctx.SendWSResponse([]*models.Player{player}, utils.GameActionJoinQueue, "matched")
		ctx.SendWSResponse(room.GetPlayers(), utils.GameActionCreateRoom, room.ID)
	} else {
		ctx.SendWSResponse([]*models.Player{player}, utils.GameActionJoinQueue, "waiting")
	}
}

func HandleCreateRoom(ctx *server.Context, player *models.Player, data any) {
	game := ctx.Game()
	room := game.CreateRoom(player)
	ctx.SendWSResponse([]*models.Player{player}, utils.GameActionCreateRoom, room.ID)
}

func HandleJoinRoom(ctx *server.Context, player *models.Player, data any) {
	roomID, ok := data.(string)
	if !ok {
		ctx.SendWSResponse([]*models.Player{player}, utils.GameActionError, "invalid_room_id")
		return
	}
	game := ctx.Game()
	room, exists := game.GetRoomByID(roomID)
	if !exists {
		ctx.SendWSResponse([]*models.Player{player}, utils.GameActionError, "room_not_found")
		return
	}
	err := room.AddPlayer(player)
	if err != nil {
		ctx.SendWSResponse([]*models.Player{player}, utils.GameActionError, err.Error())
		return
	}
	ctx.SendWSResponse([]*models.Player{player}, utils.GameActionJoinRoom, room.ID)
	ctx.SendWSResponse(room.GetPlayers(), utils.GameActionPlayerJoined, room.ID)
}
