package server

import (
	"nextduel-svr/utils"
	"os"
	"strconv"
	"strings"
	"time"
)

const tokenDir = "sessions"

func validateSession(token string) (string, bool) {
	if !validateToken(token) {
		return "", false
	}
	userID, t, valid := ParseLoginToken(token)
	if !valid {
		return "", false
	}

	if time.Since(t) > 24*time.Hour {
		handleTokenRemove(token)
		return "", false
	}

	return userID, true
}

func handleNewToken(token string) bool {
	dir := tokenDir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}

	// create token file
	tokenFile := dir + "/" + token
	f, err := os.Create(tokenFile)
	if err != nil {
		return false
	}
	f.Close()

	return true
}

func validateToken(token string) bool {
	tokenFile := tokenDir + "/" + token
	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func handleTokenRemove(token string) bool {
	tokenFile := tokenDir + "/" + token
	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		return false
	}
	return os.Remove(tokenFile) == nil
}

func NewLoginToken(userID string) string {
	builder := &strings.Builder{}
	builder.WriteString(userID)
	builder.WriteString("~")
	builder.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
	code := builder.String()
	encoded := utils.Encrypt(code)
	return encoded
}

func ParseLoginToken(token string) (string, time.Time, bool) {
	decoded := utils.Decrypt(token)
	parts := strings.Split(decoded, "~")
	if len(parts) != 2 {
		return "", time.Time{}, false
	}
	userID := parts[0]
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", time.Time{}, false
	}
	t := time.Unix(timestamp, 0)
	return userID, t, true
}

func NewGameToken(gameID string) string {
	builder := &strings.Builder{}
	builder.WriteString(gameID)
	builder.WriteString("~")
	builder.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
	encoded := utils.Encrypt(builder.String())
	return encoded
}
