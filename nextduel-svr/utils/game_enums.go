package utils

type RoomStatus string

const (
	RoomStatusCreated   RoomStatus = "created"
	RoomStatusActive    RoomStatus = "active"
	RoomStatusCompleted RoomStatus = "completed"
)

type MatchMode string

const (
	MatchModeClassic MatchMode = "classic"
)

type MatchStatus string

const (
	MatchStatusWaiting   MatchStatus = "waiting"
	MatchStatusReady     MatchStatus = "ready"
	MatchStatusActive    MatchStatus = "active"
	MatchStatusCompleted MatchStatus = "completed"
	MatchStatusAbandoned MatchStatus = "abandoned"
)

type MatchEventType string

const (
	MatchEventCreated     MatchEventType = "match_created"
	MatchEventReady       MatchEventType = "match_ready"
	MatchEventStarted     MatchEventType = "duel_start"
	MatchEventRoundStart  MatchEventType = "round_start"
	MatchEventRoundResult MatchEventType = "round_result"
	MatchEventPlayerJoin  MatchEventType = "player_joined"
	MatchEventPlayerReady MatchEventType = "player_ready"
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
