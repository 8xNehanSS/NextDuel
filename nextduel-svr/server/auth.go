package server

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"nextduel-svr/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const tokenDir = "./sessions"

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
	if err := os.MkdirAll(tokenDir, 0755); err != nil {
		log.Println("failed to create session dir:", err)
		return false
	}

	tokenFile := getTokenFilePath(token)

	f, err := os.Create(tokenFile)
	if err != nil {
		log.Println("failed to create token file:", err)
		return false
	}

	// ensure data is flushed to disk
	if err := f.Sync(); err != nil {
		log.Println("failed to sync token file:", err)
	}

	if err := f.Close(); err != nil {
		log.Println("failed to close token file:", err)
		return false
	}

	return true
}

func validateToken(token string) bool {
	tokenFile := getTokenFilePath(token)

	_, err := os.Stat(tokenFile)
	return err == nil
}

func handleTokenRemove(token string) bool {
	tokenFile := getTokenFilePath(token)

	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		return false
	}

	if err := os.Remove(tokenFile); err != nil {
		log.Println("failed to remove token file:", err)
		return false
	}

	return true
}

func NewLoginToken(userID string) string {
	builder := &strings.Builder{}
	builder.WriteString(userID)
	builder.WriteString("~")
	builder.WriteString(strconv.FormatInt(time.Now().Unix(), 10))

	code := builder.String()
	return utils.Encrypt(code)
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

	return userID, time.Unix(timestamp, 0), true
}

func NewGameToken(gameID string) string {
	builder := &strings.Builder{}
	builder.WriteString(gameID)
	builder.WriteString("~")
	builder.WriteString(strconv.FormatInt(time.Now().Unix(), 10))

	return utils.Encrypt(builder.String())
}

func getTokenFilePath(token string) string {
	return filepath.Join(tokenDir, safeTokenFilename(token))
}

func safeTokenFilename(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
