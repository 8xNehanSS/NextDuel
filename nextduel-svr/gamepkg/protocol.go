package gamepkg

import (
	"encoding/json"
	"nextduel-svr/utils"
)

// Payloads sent to clients. The TypeScript mirror lives in
// nextduel-frontend/src/types/protocol.ts — keep both in sync.

type PlayerState struct {
	ID        string `json:"id"`
	Ready     bool   `json:"ready"`
	Score     int    `json:"score"`
	RoundsWon int    `json:"roundsWon"`
}

type RoomState struct {
	RoomID    string           `json:"roomId"`
	Private   bool             `json:"private"`
	Status    utils.RoomStatus `json:"status"`
	Round     int              `json:"round"`
	MaxRounds int              `json:"maxRounds"`
	Players   []PlayerState    `json:"players"`
}

type QueuePayload struct {
	Position int `json:"position"`
}

type PlayerLeftPayload struct {
	PlayerID string `json:"playerId"`
}

type DuelStartPayload struct {
	MaxRounds int `json:"maxRounds"`
}

type RoundStartPayload struct {
	Round      int      `json:"round"`
	MaxRounds  int      `json:"maxRounds"`
	Question   string   `json:"question"`
	Category   string   `json:"category"`
	Difficulty string   `json:"difficulty"`
	Choices    []string `json:"choices"`
	Duration   int      `json:"duration"`
}

type TimerPayload struct {
	Remaining int `json:"remaining"`
}

type DuelFastPayload struct {
	PlayerID  string `json:"playerId"`
	Remaining int    `json:"remaining"`
}

type AnswerAckPayload struct {
	Round  int    `json:"round"`
	Answer string `json:"answer"`
}

type PlayerRoundResult struct {
	PlayerID  string `json:"playerId"`
	Answer    string `json:"answer"`
	Correct   bool   `json:"correct"`
	Delta     int    `json:"delta"`
	Score     int    `json:"score"`
	RoundsWon int    `json:"roundsWon"`
}

type RoundResultPayload struct {
	Round         int                 `json:"round"`
	CorrectAnswer string              `json:"correctAnswer"`
	WinnerID      *string             `json:"winnerId"`
	Results       []PlayerRoundResult `json:"results"`
}

type IntermissionPayload struct {
	Remaining int `json:"remaining"`
	NextRound int `json:"nextRound"`
}

type GameOverPayload struct {
	Reason   utils.GameOverReason `json:"reason"`
	WinnerID *string              `json:"winnerId"`
	Players  []PlayerState        `json:"players"`
}

type RoomDestroyedPayload struct {
	RoomID string                `json:"roomId"`
	Reason utils.RoomCloseReason `json:"reason"`
}

// Requests sent by clients.

type JoinRoomRequest struct {
	RoomID string `json:"roomId"`
}

type DuelAnswerRequest struct {
	Round  int    `json:"round"`
	Answer string `json:"answer"`
}

// DecodeRequest unmarshals a client payload, mapping failures to ErrBadRequest.
func DecodeRequest(data json.RawMessage, dest any) error {
	if len(data) == 0 {
		return ErrBadRequest
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return ErrBadRequest
	}
	return nil
}
