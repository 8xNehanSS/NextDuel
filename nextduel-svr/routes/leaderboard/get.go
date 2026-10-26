package leaderboard

import (
	"net/http"
	"nex-server/server"
	"nextduel-svr/repository"
	"nextduel-svr/utils"
)

const size = 25

func GetLeaderboard(ctx *server.Context) error {
	entries, err := repository.Leaderboard(ctx.DB(), utils.GameMode(ctx.GetQueryParam("mode")), size)
	if err != nil {
		return ctx.Error(http.StatusInternalServerError, "Failed to load leaderboard", err)
	}
	return ctx.Success(entries)
}
