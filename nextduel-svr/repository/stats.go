package repository

import (
	"errors"
	"nextduel-svr/models"
	"nextduel-svr/utils"
	"time"

	"gorm.io/gorm"
)

type PlayerStats struct {
	TotalMatches int     `json:"total_matches"`
	Wins         int     `json:"wins"`
	Losses       int     `json:"losses"`
	Draws        int     `json:"draws"`
	WinRate      float64 `json:"win_rate"`
}

type MatchSummary struct {
	MatchID           uint      `json:"match_id"`
	Status            string    `json:"status"`
	Result            string    `json:"result"` // win | loss | draw
	BestOf            int       `json:"best_of"`
	Score             int       `json:"score"`
	RoundsWon         int       `json:"rounds_won"`
	Opponent          string    `json:"opponent"`
	OpponentScore     int       `json:"opponent_score"`
	OpponentRoundsWon int       `json:"opponent_rounds_won"`
	StartedAt         time.Time `json:"started_at"`
	EndedAt           time.Time `json:"ended_at"`
}

type DuelHistory struct {
	Stats   PlayerStats    `json:"stats"`
	Matches []MatchSummary `json:"matches"`
	Total   int64          `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
}

var finishedStatuses = []string{string(utils.MatchStatusCompleted), string(utils.MatchStatusAbandoned)}

// History returns the finished matches of username, newest first. A user who
// has never played gets an empty history rather than an error.
func History(db *gorm.DB, username string, limit, offset int) (DuelHistory, error) {
	out := DuelHistory{Matches: []MatchSummary{}, Limit: limit, Offset: offset}

	var player models.Players
	err := db.Where("username = ?", username).First(&player).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	out.Stats = statsOf(player)

	base := db.Table("match_players AS me").
		Joins("JOIN matches m ON m.id = me.match_id").
		Where("me.player_id = ? AND m.status IN ?", player.ID, finishedStatuses)

	if err := base.Session(&gorm.Session{}).Count(&out.Total).Error; err != nil {
		return out, err
	}

	type row struct {
		MatchID           uint
		Status            string
		BestOf            int
		WinnerID          *uint
		Score             int
		RoundsWon         int
		Opponent          *string
		OpponentScore     *int
		OpponentRoundsWon *int
		StartedAt         time.Time
		EndedAt           time.Time
	}
	var rows []row
	err = base.Session(&gorm.Session{}).
		Select(`m.id AS match_id, m.status, m.best_of, m.winner_id, m.started_at, m.ended_at,
			me.score, me.rounds_won,
			opp.username AS opponent, op.score AS opponent_score, op.rounds_won AS opponent_rounds_won`).
		Joins("LEFT JOIN match_players op ON op.match_id = me.match_id AND op.player_id <> me.player_id").
		Joins("LEFT JOIN players opp ON opp.id = op.player_id").
		Order("m.started_at DESC, m.id DESC").
		Limit(limit).Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return out, err
	}

	for _, r := range rows {
		result := "draw"
		if r.WinnerID != nil {
			if *r.WinnerID == player.ID {
				result = "win"
			} else {
				result = "loss"
			}
		}
		s := MatchSummary{
			MatchID:   r.MatchID,
			Status:    r.Status,
			Result:    result,
			BestOf:    r.BestOf,
			Score:     r.Score,
			RoundsWon: r.RoundsWon,
			StartedAt: r.StartedAt,
			EndedAt:   r.EndedAt,
		}
		if r.Opponent != nil {
			s.Opponent = *r.Opponent
		}
		if r.OpponentScore != nil {
			s.OpponentScore = *r.OpponentScore
		}
		if r.OpponentRoundsWon != nil {
			s.OpponentRoundsWon = *r.OpponentRoundsWon
		}
		out.Matches = append(out.Matches, s)
	}
	return out, nil
}

type LeaderboardEntry struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
	PlayerStats
}

// Leaderboard ranks players by wins, then win rate, then fewest matches.
func Leaderboard(db *gorm.DB, limit int) ([]LeaderboardEntry, error) {
	var players []models.Players
	err := db.Where("total_matches > 0").
		Order("total_wins DESC").
		Order("CAST(total_wins AS REAL) / total_matches DESC").
		Order("total_matches ASC").
		Order("username ASC").
		Limit(limit).
		Find(&players).Error
	if err != nil {
		return nil, err
	}

	out := make([]LeaderboardEntry, len(players))
	for i, p := range players {
		out[i] = LeaderboardEntry{Rank: i + 1, Username: p.Username, PlayerStats: statsOf(p)}
	}
	return out, nil
}

func statsOf(p models.Players) PlayerStats {
	s := PlayerStats{
		TotalMatches: p.TotalMatches,
		Wins:         p.TotalWins,
		Losses:       p.TotalLosses,
		Draws:        max(p.TotalMatches-p.TotalWins-p.TotalLosses, 0),
	}
	if s.TotalMatches > 0 {
		s.WinRate = float64(s.Wins) / float64(s.TotalMatches)
	}
	return s
}
