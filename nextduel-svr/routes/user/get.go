package user

import (
	"net/http"
	"nextduel-svr/models"
	"nextduel-svr/server"
)

func GetUser(ctx *server.Context) error {
	userID := ctx.GetCurrentUserID()
	db := ctx.DB()
	var user models.User
	if err := db.First(&user, "username = ?", userID).Error; err != nil {
		return ctx.Error(http.StatusNotFound, "user not found", err)
	}
	return ctx.Success(user)
}
