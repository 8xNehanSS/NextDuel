package health

import (
	"nex-server/server"
)

func (h *HealthServices) HealthHandler(ctx *server.Context) error {
	if h.Game == nil {
		ctx.InternalError("game service unavailable")
		return nil
	}
	ctx.JSON(200, h.Game.GetStatistics())
	return nil
}
