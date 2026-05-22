package game_events

import (
	"nextduel-svr/models"
	"nextduel-svr/server"
)

func HandleJoinQueue(ctx *server.Context, player *models.Player, data any) {
	game := ctx.Game()
	matched := game.AddToQueue(player)
	if matched {
		ctx.SendWSResponse([]*models.Player{player}, "joined_queue", "matched")
	} else {
		ctx.SendWSResponse([]*models.Player{player}, "joined_queue", "waiting")
	}
}

func HandleCreateRoom(ctx *server.Context, player *models.Player, data any) {
	game := ctx.Game()
	room := game.CreateRoom(player)
	ctx.SendWSResponse([]*models.Player{player}, "created_room", room.ID)
}

func HandleJoinRoom(ctx *server.Context, player *models.Player, data any) {
	roomID, ok := data.(string)
	if !ok {
		ctx.SendWSResponse([]*models.Player{player}, "error", "invalid_room_id")
		return
	}
	game := ctx.Game()
	room, exists := game.GetRoomByID(roomID)
	if !exists {
		ctx.SendWSResponse([]*models.Player{player}, "error", "room_not_found")
		return
	}
	err := room.AddPlayer(player)
	if err != nil {
		ctx.SendWSResponse([]*models.Player{player}, "error", err.Error())
		return
	}
	ctx.SendWSResponse([]*models.Player{player}, "joined_room", room.ID)
	ctx.SendWSResponse(room.GetPlayers(), "player_joined", room.ID)
}
