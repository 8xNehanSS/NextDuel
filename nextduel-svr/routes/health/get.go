package health

import (
	"context"
	"net/http"
	"nex-server/server"
	"time"
)

type HealthResponse struct {
	Status   string `json:"status"`
	Version  string `json:"version"`
	Database string `json:"database"`
}

// HealthHandler is a public liveness/readiness probe.
func (h *HealthServices) HealthHandler(ctx *server.Context) error {
	resp := HealthResponse{Status: "ok", Version: h.Version, Database: "ok"}

	pingCtx, cancel := context.WithTimeout(ctx.Request().Context(), 2*time.Second)
	defer cancel()
	sqlDB, err := ctx.DB().DB()
	if err == nil {
		err = sqlDB.PingContext(pingCtx)
	}
	if err != nil {
		resp.Status, resp.Database = "degraded", "unreachable"
		return ctx.JSON(http.StatusServiceUnavailable, resp)
	}
	return ctx.OK(resp)
}

// StatsHandler reports live game and process statistics (authenticated).
func (h *HealthServices) StatsHandler(ctx *server.Context) error {
	if h.Game == nil {
		return ctx.InternalError("game service unavailable")
	}
	return ctx.Success(h.Game.Statistics())
}
