package game_events

import (
	"nextduel-svr/gamepkg"
	"nextduel-svr/server"
	"nextduel-svr/utils"
)

func HandleJoinQueue(ctx *server.Context, player *gamepkg.Player, data any) {
	game := ctx.Game()
	room, matched, err := game.AddToQueue(player)
	if err != nil {
		ctx.SendWSResponse([]*gamepkg.Player{player}, utils.GameActionError, err.Error())
		return
	}
	if matched {
		ctx.SendWSResponse(room.GetPlayers(), utils.GameActionMatchFound, nil)
		ctx.SendWSResponse(room.GetPlayers(), utils.GameActionRoomCreated, room.ID)
	} else {
		ctx.SendWSResponse([]*gamepkg.Player{player}, utils.GameActionWaitingQueue, nil)
	}
}

func HandleCreateRoom(ctx *server.Context, player *gamepkg.Player, data any) {
	game := ctx.Game()
	room := game.CreateRoom(player)
	ctx.SendWSResponse([]*gamepkg.Player{player}, utils.GameActionRoomCreated, room.ID)
}

func HandleJoinRoom(ctx *server.Context, player *gamepkg.Player, data any) {
	roomID, ok := data.(string)
	if !ok {
		ctx.SendWSResponse([]*gamepkg.Player{player}, utils.GameActionError, "invalid_room_id")
		return
	}
	game := ctx.Game()
	room, exists := game.GetRoomByID(roomID)
	if !exists {
		ctx.SendWSResponse([]*gamepkg.Player{player}, utils.GameActionError, "room_not_found")
		return
	}
	err := room.AddPlayer(player)
	if err != nil {
		ctx.SendWSResponse([]*gamepkg.Player{player}, utils.GameActionError, err.Error())
		return
	}
	ctx.SendWSResponse([]*gamepkg.Player{player}, utils.GameActionRoomJoined, room.ID)
	ctx.SendWSResponse(room.GetPlayers(), utils.GameActionPlayerJoined, room.ID)
}
