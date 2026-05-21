package health

import "nextduel-svr/server"

func HealthHandler(ctx *server.Context) error {
	ctx.JSON(200, map[string]string{"status": "ok"})
	return nil
}
