package user

import (
	"net/http"
	"nex-server/server"
	"nextduel-svr/models"
)

func GetUser(ctx *server.Context) error {
	userID := ctx.UserID()
	db := ctx.DB()
	var user models.User
	if err := db.First(&user, "username = ?", userID).Error; err != nil {
		return ctx.Error(http.StatusNotFound, "user not found", err)
	}
	return ctx.Success(user)
}
