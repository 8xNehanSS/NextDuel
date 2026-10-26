package history

import (
	"net/http"
	"nex-server/server"
	"nextduel-svr/repository"
	"strconv"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

func GetDuelHistory(ctx *server.Context) error {
	limit := intParam(ctx, "limit", defaultLimit)
	if limit < 1 || limit > maxLimit {
		limit = defaultLimit
	}
	offset := max(intParam(ctx, "offset", 0), 0)

	history, err := repository.History(ctx.DB(), ctx.UserID(), limit, offset)
	if err != nil {
		return ctx.Error(http.StatusInternalServerError, "Failed to load duel history", err)
	}
	return ctx.Success(history)
}

func intParam(ctx *server.Context, name string, fallback int) int {
	v, err := strconv.Atoi(ctx.GetQueryParam(name))
	if err != nil {
		return fallback
	}
	return v
}
