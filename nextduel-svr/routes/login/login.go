package login

import (
	"errors"
	"net/http"
	"nex-server/server"
	serverutils "nex-server/utils"
	"nextduel-svr/models"
	"strings"
	"sync"

	"gorm.io/gorm"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Username string `json:"username"`
}

var (
	dummyHashOnce sync.Once
	dummyHash     string
)

// Comparing against a dummy hash when the user does not exist keeps response
// times uniform, so usernames cannot be enumerated by timing.
func compareDummy(password string) {
	dummyHashOnce.Do(func() {
		dummyHash, _ = serverutils.HashPassword("nextduel-timing-equaliser")
	})
	serverutils.CheckPassword(dummyHash, password)
}

func LoginHandler(ctx *server.Context) error {
	var req LoginRequest
	if err := ctx.GetRequestBodyAs(&req); err != nil {
		return ctx.BadRequest("Invalid request body")
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		return ctx.BadRequest("Username and password are required")
	}

	var user models.User
	err := ctx.DB().Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Error(http.StatusInternalServerError, "Login failed", err)
		}
		compareDummy(req.Password)
		return ctx.Unauthorized("Invalid username or password")
	}

	if !serverutils.CheckPassword(user.Password, req.Password) {
		return ctx.Unauthorized("Invalid username or password")
	}
	if !user.IsActive {
		return ctx.Forbidden("This account has been disabled")
	}

	ctx.LoginSuccess(user.Username)
	return ctx.Success(LoginResponse{Username: user.Username})
}
