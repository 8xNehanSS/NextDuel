package health

import (
	"nex-server/server"
	"nextduel-svr/gamepkg"
)

func HealthHandler(ctx *server.Context) error {
	game, ok := server.Get[*gamepkg.GameManager](ctx)
	if !ok {
		ctx.InternalError("game service unavailable")
		return nil
	}
	ctx.JSON(200, game.GetStatistics())
	return nil
}
