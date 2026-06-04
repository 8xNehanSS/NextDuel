package history

import (
	"nex-server/server"
	"nextduel-svr/models"
)

func GetDuelHistory(ctx *server.Context) error {
	// select matches where mode = 'duel' and status = 'completed'
	db := ctx.DB()
	// get player record first
	var player models.Players
	row := db.Model(&models.Players{}).Where("username = ?", ctx.UserID()).First(&player)
	if row.Error != nil {
		ctx.Error(500, "Failed to find player record", row.Error)
		return nil
	}
	// then get match records
	rows := db.Model(&models.MatchPlayer{}).Where("player_id = ?", player.ID)
	var matchPlayers []models.MatchPlayer
	if err := rows.Find(&matchPlayers).Error; err != nil {
		ctx.Error(500, "Failed to fetch duel history", err)
		return nil
	}
	ctx.JSON(200, matchPlayers)
	return nil
}
