package health

import "nextduel-svr/server"

func HealthHandler(ctx *server.Context) error {
	ctx.JSON(200, ctx.Game().GetStatistics())
	return nil
}
