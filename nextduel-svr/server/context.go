package server

import (
	"encoding/json"
	"log"
	"net/http"
	"nextduel-svr/game"
	"nextduel-svr/models"
	"nextduel-svr/utils"

	"gorm.io/gorm"
)

type Context struct {
	writer  http.ResponseWriter
	request *http.Request

	userID     string
	authorized bool

	db     *gorm.DB
	Config utils.Config

	// game related
	gameManager *game.GameManager
}

func NewContext(w http.ResponseWriter, r *http.Request, userID string, authorized bool, db *gorm.DB, config utils.Config, gameManager *game.GameManager) *Context {
	return &Context{
		writer:      w,
		request:     r,
		userID:      userID,
		authorized:  authorized,
		db:          db,
		Config:      config,
		gameManager: gameManager,
	}
}

func NewUnauthenticatedContext(w http.ResponseWriter, r *http.Request, db *gorm.DB, config utils.Config, gameManager *game.GameManager) *Context {
	return &Context{
		writer:      w,
		request:     r,
		userID:      "",
		authorized:  false,
		db:          db,
		Config:      config,
		gameManager: gameManager,
	}
}

func (ctx *Context) Game() *game.GameManager {
	return ctx.gameManager
}

func (ctx *Context) DB() *gorm.DB {
	return ctx.db
}

func (ctx *Context) GetCurrentUserID() string {
	return ctx.userID
}

func (ctx *Context) IsAuthorized() bool {
	return ctx.authorized
}

func (ctx *Context) SetAuthenticated(userID string) {
	ctx.userID = userID
	ctx.authorized = true
}

func (ctx *Context) ClearAuthentication() {
	ctx.userID = ""
	ctx.authorized = false
}

func (ctx *Context) LoginSuccess(userID string) {
	ctx.SetAuthenticated(userID)
	token := NewLoginToken(userID)
	handleNewToken(token)
	ctx.setCookie("NEXT-DUEL", token)
}

func (ctx *Context) Logout() {
	token, err := ctx.request.Cookie("NEXT-DUEL")
	if err == nil && token.Value != "" {
		handleTokenRemove(token.Value)
		ctx.setCookie("NEXT-DUEL", "")
	}
}

func (ctx *Context) setCookie(name, value string) {
	http.SetCookie(ctx.writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	})
}

func (ctx *Context) SendWSResponse(player []*models.Player, action utils.GameAction, data any) {
	for _, p := range player {
		err := p.Conn.WriteJSON(map[string]any{
			"type": action,
			"data": data,
		})

		if err != nil {
			log.Println("Write error:", err)
		}
	}
}

func (ctx *Context) JSON(status int, data any) error {
	ctx.writer.Header().Set("Content-Type", "application/json")
	ctx.writer.WriteHeader(status)
	return json.NewEncoder(ctx.writer).Encode(data)
}

func (ctx *Context) Text(status int, text string) error {
	ctx.writer.Header().Set("Content-Type", "text/plain")
	ctx.writer.WriteHeader(status)
	_, err := ctx.writer.Write([]byte(text))
	return err
}

func (ctx *Context) WriteFile(filePath string) {
	http.ServeFile(ctx.writer, ctx.request, filePath)
}

func (ctx *Context) OK(data any) error {
	return ctx.JSON(http.StatusOK, data)
}

func (ctx *Context) Success(data any) error {
	return ctx.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
	})
}

func (ctx *Context) Created(data any) error {
	return ctx.JSON(http.StatusCreated, map[string]any{
		"success": true,
		"data":    data,
	})
}

func (ctx *Context) NoContent() error {
	ctx.writer.WriteHeader(http.StatusNoContent)
	return nil
}

func (ctx *Context) Error(status int, message string, err error) error {
	log.Printf("Error: %s - %v", message, err)
	return ctx.JSON(status, map[string]any{
		"success": false,
		"error":   message,
		"stack":   err.Error(),
	})
}

func (ctx *Context) Fail(message string) error {
	return ctx.JSON(http.StatusOK, map[string]any{
		"success": false,
		"error":   message,
	})
}

func (ctx *Context) BadRequest(message string) error {
	return ctx.Error(http.StatusBadRequest, message, nil)
}

func (ctx *Context) Unauthorized(message string) error {
	if message == "" {
		message = "unauthorized"
	}
	return ctx.Error(http.StatusUnauthorized, message, nil)
}

func (ctx *Context) Forbidden(message string) error {
	if message == "" {
		message = "forbidden"
	}
	return ctx.Error(http.StatusForbidden, message, nil)
}

func (ctx *Context) NotFound(message string) error {
	if message == "" {
		message = "not found"
	}
	return ctx.Error(http.StatusNotFound, message, nil)
}

func (ctx *Context) InternalError(message string) error {
	if message == "" {
		message = "internal server error"
	}
	return ctx.Error(http.StatusInternalServerError, message, nil)
}

func (ctx *Context) GetRequestBodyAs(dest any) error {
	return json.NewDecoder(ctx.request.Body).Decode(dest)
}

func (ctx *Context) GetQueryParam(param string) string {
	return ctx.request.URL.Query().Get(param)
}

func (ctx *Context) GetFormValue(key string) string {
	return ctx.request.FormValue(key)
}

func (ctx *Context) GetHeader(key string) string {
	return ctx.request.Header.Get(key)
}

func (ctx *Context) SetHeader(key, value string) {
	ctx.writer.Header().Set(key, value)
}

func (ctx *Context) Redirect(url string) error {
	return ctx.JSON(http.StatusSeeOther, map[string]any{
		"success":  true,
		"redirect": url,
	})
}

func (ctx *Context) Request() *http.Request {
	return ctx.request
}

func (ctx *Context) ResponseWriter() http.ResponseWriter {
	return ctx.writer
}
