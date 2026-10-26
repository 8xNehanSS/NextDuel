package register

import (
	"net/http"
	"net/mail"
	"nex-server/server"
	serverutils "nex-server/utils"
	"nextduel-svr/models"
	"regexp"
	"strings"
	"unicode/utf8"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,20}$`)

const (
	minPasswordLength = 8
	// bcrypt ignores input past 72 bytes.
	maxPasswordBytes = 72
)

// Validate normalises the request and returns a user-facing error message.
func (req *RegisterRequest) Validate() string {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if !usernamePattern.MatchString(req.Username) {
		return "Username must be 3-20 characters: letters, numbers or underscores"
	}
	if addr, err := mail.ParseAddress(req.Email); err != nil || addr.Address != req.Email || len(req.Email) > 100 {
		return "Please enter a valid email address"
	}
	if utf8.RuneCountInString(req.Password) < minPasswordLength {
		return "Password must be at least 8 characters"
	}
	if len(req.Password) > maxPasswordBytes {
		return "Password must be at most 72 bytes"
	}
	return ""
}

func RegisterUser(ctx *server.Context) error {
	var req RegisterRequest
	if err := ctx.GetRequestBodyAs(&req); err != nil {
		return ctx.BadRequest("Invalid request body")
	}
	if msg := req.Validate(); msg != "" {
		return ctx.BadRequest(msg)
	}

	db := ctx.DB()
	var count int64
	if err := db.Model(&models.User{}).Where("LOWER(username) = LOWER(?)", req.Username).Count(&count).Error; err != nil {
		return ctx.Error(http.StatusInternalServerError, "Registration failed", err)
	}
	if count > 0 {
		return ctx.Error(http.StatusConflict, "That username is taken", nil)
	}
	if err := db.Model(&models.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		return ctx.Error(http.StatusInternalServerError, "Registration failed", err)
	}
	if count > 0 {
		return ctx.Error(http.StatusConflict, "An account with that email already exists", nil)
	}

	hashedPassword, err := serverutils.HashPassword(req.Password)
	if err != nil {
		return ctx.Error(http.StatusInternalServerError, "Registration failed", err)
	}

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "user",
		IsActive: true,
	}
	if err := db.Create(&user).Error; err != nil {
		// Unique indexes catch the race between the checks above and here.
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return ctx.Error(http.StatusConflict, "That username or email is already registered", nil)
		}
		return ctx.Error(http.StatusInternalServerError, "Registration failed", err)
	}

	return ctx.Created(user)
}
