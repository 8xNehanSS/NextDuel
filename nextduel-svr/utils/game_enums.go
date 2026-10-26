package utils

type RoomStatus string

const (
	// RoomStatusWaiting: waiting for a second player and/or ready-ups.
	RoomStatusWaiting RoomStatus = "waiting"
	// RoomStatusStarting: everyone is ready, questions are being fetched.
	RoomStatusStarting  RoomStatus = "starting"
	RoomStatusActive    RoomStatus = "active"
	RoomStatusCompleted RoomStatus = "completed"
)

type MatchMode string

const (
	MatchModeClassic MatchMode = "classic"
)

type MatchStatus string

const (
	MatchStatusActive    MatchStatus = "active"
	MatchStatusCompleted MatchStatus = "completed"
	// MatchStatusAbandoned: a player left mid-match; the other wins.
	MatchStatusAbandoned MatchStatus = "abandoned"
	// MatchStatusAborted: ended with no result (e.g. server shutdown).
	MatchStatusAborted MatchStatus = "aborted"
)

type MatchEventType string

const (
	MatchEventCreated     MatchEventType = "match_created"
	MatchEventStarted     MatchEventType = "duel_start"
	MatchEventRoundStart  MatchEventType = "round_start"
	MatchEventRoundResult MatchEventType = "round_result"
	MatchEventPlayerJoin  MatchEventType = "player_joined"
	MatchEventPlayerLeft  MatchEventType = "player_left"
	MatchEventDuelAnswer  MatchEventType = "duel_answer"
	MatchEventDuelDone    MatchEventType = "duel_complete"
)

type ScoreEventType string

const (
	ScoreEventRoundScore ScoreEventType = "round_score"
)

type ScoreEventReason string

const (
	ScoreReasonRoundEnd   ScoreEventReason = "round_end"
	ScoreReasonDisconnect ScoreEventReason = "disconnect"
)

// GameOverReason explains why a duel ended.
type GameOverReason string

const (
	GameOverCompleted GameOverReason = "completed"
	GameOverForfeit   GameOverReason = "forfeit"
)

// RoomCloseReason is sent with room_destroyed when a room closes before a
// result could be produced.
type RoomCloseReason string

const (
	RoomCloseOpponentLeft   RoomCloseReason = "opponent_left"
	RoomCloseNoQuestions    RoomCloseReason = "questions_unavailable"
	RoomCloseServerShutdown RoomCloseReason = "server_shutdown"
)
