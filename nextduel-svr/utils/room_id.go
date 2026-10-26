package utils

import (
	"crypto/rand"
	"math/big"
)

// Unambiguous alphabet (no 0/O, 1/I/L) so codes are easy to read aloud.
const roomCodeCharset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

const RoomCodeLength = 6

// GenerateRoomID returns a short, human-shareable room code.
// Uniqueness against live rooms is enforced by the caller.
func GenerateRoomID() string {
	b := make([]byte, RoomCodeLength)
	max := big.NewInt(int64(len(roomCodeCharset)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic("crypto/rand unavailable: " + err.Error())
		}
		b[i] = roomCodeCharset[n.Int64()]
	}
	return string(b)
}
