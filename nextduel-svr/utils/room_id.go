package utils

import (
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRoomID() string {
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return "room_" + string(b)
}
