package login

import (
	"net/http"
	"nex-server/server"
	serverutils "nex-server/utils"
	"nextduel-svr/models"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginHandler(ctx *server.Context) error {
	var req LoginRequest
	err := ctx.GetRequestBodyAs(&req)
	if err != nil {
		return ctx.Error(http.StatusBadRequest, "Invalid request body", err)
	}

	var user models.User
	err = ctx.DB().Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		return ctx.Error(http.StatusUnauthorized, "Invalid username or password", err)
	}

	if !serverutils.CheckPassword(user.Password, req.Password) {
		return ctx.Error(http.StatusUnauthorized, "Invalid username or password", err)
	}

	ctx.LoginSuccess(user.Username)
	redirect := ctx.GetQueryParam("redirect")
	if redirect == "" {
		redirect = "/"
	}
	ctx.Redirect(redirect)
	return nil
}
