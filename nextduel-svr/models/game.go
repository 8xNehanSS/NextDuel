package models

import (
	"time"
)

type Players struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Username     string `gorm:"uniqueIndex;size:50;not null" json:"username"`
	TotalMatches int    `gorm:"default:0" json:"total_matches"`
	TotalWins    int    `gorm:"default:0" json:"total_wins"`
	TotalLosses  int    `gorm:"default:0" json:"total_losses"`

	CreatedAt time.Time `json:"created_at"`
}

type Match struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Mode      string `gorm:"size:20;default:'trivia';index" json:"mode"`
	Status    string `gorm:"size:20;default:'waiting';index" json:"status"`
	BestOf    int    `gorm:"default:3" json:"best_of"`
	WinnerID  *uint  `gorm:"index" json:"winner_id"`
	CreatedBy uint   `gorm:"index" json:"created_by"`

	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MatchPlayer struct {
	ID uint `gorm:"primaryKey" json:"id"`

	MatchID  uint `gorm:"index:idx_match_player,unique" json:"match_id"`
	PlayerID uint `gorm:"index:idx_match_player,unique;index" json:"player_id"`
	Slot     int  `gorm:"default:1" json:"slot"`

	RoundsWon int `gorm:"default:0" json:"rounds_won"`
	Score     int `gorm:"default:0" json:"score"`

	JoinedAt time.Time `json:"joined_at"`
	LeftAt   time.Time `json:"left_at"`
}

type MatchRound struct {
	ID uint `gorm:"primaryKey" json:"id"`

	MatchID  uint  `gorm:"index:idx_match_round,unique" json:"match_id"`
	RoundNo  int   `gorm:"index:idx_match_round,unique" json:"round_no"`
	WinnerID *uint `gorm:"index" json:"winner_id"`

	QuestID  string `gorm:"size:64;index" json:"quest_id"`
	Question string `gorm:"type:text" json:"question"`

	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}

type RoundPlayer struct {
	ID uint `gorm:"primaryKey" json:"id"`

	RoundID  uint `gorm:"index:idx_round_player,unique" json:"round_id"`
	PlayerID uint `gorm:"index:idx_round_player,unique;index" json:"player_id"`

	ScoreDelta int       `gorm:"default:0" json:"score_delta"`
	IsWinner   bool      `gorm:"default:false" json:"is_winner"`
	Answer     string    `gorm:"type:text" json:"answer"`
	AnsweredAt time.Time `json:"answered_at"`
}

type ScoreEvent struct {
	ID uint `gorm:"primaryKey" json:"id"`

	MatchID  uint  `gorm:"index" json:"match_id"`
	RoundID  *uint `gorm:"index" json:"round_id"`
	PlayerID uint  `gorm:"index" json:"player_id"`

	EventType string `gorm:"size:50;index" json:"event_type"`
	Reason    string `gorm:"size:100" json:"reason"`
	Delta     int    `json:"delta"`
	Total     int    `json:"total"`

	CreatedAt time.Time `json:"created_at"`
}

type MatchEventLog struct {
	ID uint `gorm:"primaryKey" json:"id"`

	MatchID  uint  `gorm:"index" json:"match_id"`
	RoundID  *uint `gorm:"index" json:"round_id"`
	PlayerID *uint `gorm:"index" json:"player_id"`

	EventType  string    `gorm:"size:50;index" json:"event_type"`
	Payload    string    `gorm:"type:text" json:"payload"`
	OccurredAt time.Time `gorm:"index" json:"occurred_at"`
}
