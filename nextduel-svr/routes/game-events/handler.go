// Package game_events adapts websocket messages onto the GameManager.
package game_events

import (
	"encoding/json"
	"log/slog"
	"nex-server/server"
	"nextduel-svr/gamepkg"
	"nextduel-svr/utils"
	"runtime/debug"

	"github.com/gorilla/websocket"
)

type GameEventServices struct {
	Game *gamepkg.GameManager
}

func New(game *gamepkg.GameManager) *GameEventServices {
	return &GameEventServices{
		Game: game,
	}
}

type eventHandler func(g *gamepkg.GameManager, p *gamepkg.Player, data json.RawMessage) error

var handlers = map[utils.GameAction]eventHandler{
	utils.GameActionPing: func(_ *gamepkg.GameManager, p *gamepkg.Player, _ json.RawMessage) error {
		p.Send(utils.GameActionPong, nil)
		return nil
	},
	utils.GameActionJoinQueue: func(g *gamepkg.GameManager, p *gamepkg.Player, data json.RawMessage) error {
		mode, err := decodeMode(data)
		if err != nil {
			return err
		}
		return g.JoinQueue(p, mode)
	},
	utils.GameActionLeaveQueue: func(g *gamepkg.GameManager, p *gamepkg.Player, _ json.RawMessage) error {
		return g.LeaveQueue(p)
	},
	utils.GameActionCreateRoom: func(g *gamepkg.GameManager, p *gamepkg.Player, data json.RawMessage) error {
		mode, err := decodeMode(data)
		if err != nil {
			return err
		}
		return g.CreateRoom(p, mode)
	},
	utils.GameActionJoinRoom: func(g *gamepkg.GameManager, p *gamepkg.Player, data json.RawMessage) error {
		var req gamepkg.JoinRoomRequest
		if err := gamepkg.DecodeRequest(data, &req); err != nil {
			return err
		}
		return g.JoinRoom(p, req.RoomID)
	},
	utils.GameActionLeaveRoom: func(g *gamepkg.GameManager, p *gamepkg.Player, _ json.RawMessage) error {
		return g.LeaveRoom(p)
	},
	utils.GameActionPlayerReady: func(g *gamepkg.GameManager, p *gamepkg.Player, _ json.RawMessage) error {
		return g.SetReady(p)
	},
	utils.GameActionDuelAnswer: func(g *gamepkg.GameManager, p *gamepkg.Player, data json.RawMessage) error {
		var req gamepkg.DuelAnswerRequest
		if err := gamepkg.DecodeRequest(data, &req); err != nil {
			return err
		}
		return g.SubmitAnswer(p, req.Round, req.Answer)
	},
}

// HandleMessage is the nex-server WSHandler. It never returns an error:
// returning one would make nex-server close the connection, and a bad message
// from the client is not a reason to drop the session.
func (g *GameEventServices) HandleMessage(ctx *server.Context, conn *websocket.Conn, msg server.WSMessage, _ any) error {
	player := g.Game.Session(ctx.UserID(), conn)

	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("websocket handler panicked", "type", msg.Type, "player", player.ID, "panic", rec, "stack", string(debug.Stack()))
			player.SendError(nil)
		}
	}()

	handler, ok := handlers[utils.GameAction(msg.Type)]
	if !ok {
		slog.Debug("unknown websocket message", "type", msg.Type, "player", player.ID)
		player.SendError(gamepkg.ErrUnknownAction)
		return nil
	}
	if err := handler(g.Game, player, msg.Data); err != nil {
		player.SendError(err)
	}
	return nil
}

// HandleDisconnect is the nex-server WSEndHandler.
func (g *GameEventServices) HandleDisconnect(_ *server.Context, conn *websocket.Conn, _ any) error {
	g.Game.EndSession(conn)
	return nil
}

// decodeMode reads an optional {mode} payload; absent or null means default.
func decodeMode(data json.RawMessage) (utils.GameMode, error) {
	if len(data) == 0 || string(data) == "null" {
		return "", nil
	}
	var req gamepkg.ModeRequest
	if err := gamepkg.DecodeRequest(data, &req); err != nil {
		return "", err
	}
	return req.Mode, nil
}
