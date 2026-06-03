package register

import (
	"net/http"
	"nex-server/server"
	serverutils "nex-server/utils"
	"nextduel-svr/models"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func RegisterUser(ctx *server.Context) error {
	var req RegisterRequest
	err := ctx.GetRequestBodyAs(&req)
	if err != nil {
		return ctx.Error(http.StatusBadRequest, "Invalid request body", err)
	}

	if !verifyUsername(ctx, req.Username) {
		return ctx.Error(http.StatusBadRequest, "Username already exists", err)
	}

	if !verifyEmail(ctx, req.Email) {
		return ctx.Error(http.StatusBadRequest, "Email already exists", err)
	}

	hashedPassword, err := serverutils.HashPassword(req.Password)
	if err != nil {
		return ctx.Error(http.StatusInternalServerError, "Failed to hash password", err)
	}

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "user",
		IsActive: false,
	}

	if err := ctx.DB().Create(&user).Error; err != nil {
		return ctx.Error(http.StatusInternalServerError, "Failed to create user", err)
	}

	ctx.Created(user)
	return nil
}

func verifyUsername(ctx *server.Context, username string) bool {
	var count int64
	ctx.DB().Model(&models.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return false
	}
	return true
}

func verifyEmail(ctx *server.Context, email string) bool {
	var count int64
	ctx.DB().Model(&models.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		return false
	}
	return true
}
