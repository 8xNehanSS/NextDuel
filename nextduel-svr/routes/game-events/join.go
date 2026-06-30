package game_events

import (
	"nex-server/server"
	"nextduel-svr/gamepkg"
	"nextduel-svr/utils"
)

func (g *GameEventServices) HandleJoinQueue(ctx *server.Context, player *gamepkg.Player, data any) {
	if g.Game == nil {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "game service not found")
		return
	}
	room, matched, err := g.Game.AddToQueue(player)
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

func (g *GameEventServices) HandleCreateRoom(ctx *server.Context, player *gamepkg.Player, data any) {
	if g.Game == nil {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "game service not found")
		return
	}
	room := g.Game.CreateRoom(player)
	ctx.SendWSResponse(player.Conn, utils.GameActionRoomCreated, room.ID)
}

func (g *GameEventServices) HandleJoinRoom(ctx *server.Context, player *gamepkg.Player, data any) {
	roomID, ok := data.(string)
	if !ok {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "invalid_room_id")
		return
	}
	if g.Game == nil {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "game service not found")
		return
	}
	if !ok {
		ctx.SendWSResponse(player.Conn, utils.GameActionError, "game service not found")
		return
	}
	room, exists := g.Game.GetRoomByID(roomID)
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
