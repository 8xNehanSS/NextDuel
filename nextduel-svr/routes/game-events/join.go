package game_events

import (
	"nex-server/server"
	"nextduel-svr/gamepkg"
	"nextduel-svr/utils"
)

func HandleJoinQueue(ctx *server.Context, player *gamepkg.Player, data any) {
	game, ok := server.Get[*gamepkg.GameManager](ctx)
	if !ok {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "game service not found")
		return
	}
	room, matched, err := game.AddToQueue(player)
	if err != nil {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, err.Error())
		return
	}
	if matched {
		ctx.SendMultipleWSResponses(room.GetPlayerCons(), utils.GameActionMatchFound, nil)
		ctx.SendMultipleWSResponses(room.GetPlayerCons(), utils.GameActionRoomCreated, room.ID)
	} else {
		ctx.SendWSResponse(player.Conn, utils.GameActionWaitingQueue, nil)
	}
}

func HandleCreateRoom(ctx *server.Context, player *gamepkg.Player, data any) {
	game, ok := server.Get[*gamepkg.GameManager](ctx)
	if !ok {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "game service not found")
		return
	}
	room := game.CreateRoom(player)
	ctx.SendWSResponse(player.Conn, utils.GameActionRoomCreated, room.ID)
}

func HandleJoinRoom(ctx *server.Context, player *gamepkg.Player, data any) {
	roomID, ok := data.(string)
	if !ok {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "invalid_room_id")
		return
	}
	game, ok := server.Get[*gamepkg.GameManager](ctx)
	if !ok {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "game service not found")
		return
	}
	room, exists := game.GetRoomByID(roomID)
	if !exists {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "room_not_found")
		return
	}
	err := room.AddPlayer(player)
	if err != nil {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, err.Error())
		return
	}
	ctx.SendWSResponse(player.Conn, utils.GameActionRoomJoined, room.ID)
	ctx.SendMultipleWSResponses(room.GetPlayerCons(), utils.GameActionPlayerJoined, room.ID)
}
