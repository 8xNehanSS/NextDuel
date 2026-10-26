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
	Mode              string    `json:"mode"`
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

// History returns the finished matches of username, newest first, optionally
// limited to one game mode (mode == "" means all). Stats cover the same
// selection. A user who has never played gets an empty history, not an error.
func History(db *gorm.DB, username string, mode utils.GameMode, limit, offset int) (DuelHistory, error) {
	out := DuelHistory{Matches: []MatchSummary{}, Limit: limit, Offset: offset}

	var player models.Players
	err := db.Where("username = ?", username).First(&player).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	base := db.Table("match_players AS me").
		Joins("JOIN matches m ON m.id = me.match_id").
		Where("me.player_id = ? AND m.status IN ?", player.ID, finishedStatuses)
	if mode != "" {
		base = base.Where("m.mode = ?", string(mode))
		if out.Stats, err = modeStats(db, player.ID, mode); err != nil {
			return out, err
		}
	} else {
		out.Stats = statsOf(player)
	}

	if err := base.Session(&gorm.Session{}).Count(&out.Total).Error; err != nil {
		return out, err
	}

	type row struct {
		MatchID           uint
		Mode              string
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
		Select(`m.id AS match_id, m.mode, m.status, m.best_of, m.winner_id, m.started_at, m.ended_at,
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
			Mode:      r.Mode,
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
// mode == "" ranks lifetime totals across every game; otherwise only matches
// of that mode count.
func Leaderboard(db *gorm.DB, mode utils.GameMode, limit int) ([]LeaderboardEntry, error) {
	type row struct {
		Username     string
		TotalMatches int
		TotalWins    int
		TotalLosses  int
	}
	var rows []row
	var q *gorm.DB
	if mode == "" {
		q = db.Model(&models.Players{}).
			Select("username, total_matches, total_wins, total_losses").
			Where("total_matches > 0")
	} else {
		q = db.Table("match_players AS mp").
			Select(`p.username,
				COUNT(*) AS total_matches,
				SUM(CASE WHEN m.winner_id = mp.player_id THEN 1 ELSE 0 END) AS total_wins,
				SUM(CASE WHEN m.winner_id IS NOT NULL AND m.winner_id <> mp.player_id THEN 1 ELSE 0 END) AS total_losses`).
			Joins("JOIN matches m ON m.id = mp.match_id").
			Joins("JOIN players p ON p.id = mp.player_id").
			Where("m.status IN ? AND m.mode = ?", finishedStatuses, string(mode)).
			Group("p.id, p.username")
	}
	err := q.Order("total_wins DESC").
		Order("CAST(total_wins AS REAL) / total_matches DESC").
		Order("total_matches ASC").
		Order("username ASC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]LeaderboardEntry, len(rows))
	for i, r := range rows {
		out[i] = LeaderboardEntry{Rank: i + 1, Username: r.Username, PlayerStats: statsOf(models.Players{
			TotalMatches: r.TotalMatches, TotalWins: r.TotalWins, TotalLosses: r.TotalLosses,
		})}
	}
	return out, nil
}

// modeStats aggregates one player's results for a single game mode.
func modeStats(db *gorm.DB, playerID uint, mode utils.GameMode) (PlayerStats, error) {
	var r struct {
		TotalMatches int
		TotalWins    int
		TotalLosses  int
	}
	err := db.Table("match_players AS mp").
		Select(`COUNT(*) AS total_matches,
			COALESCE(SUM(CASE WHEN m.winner_id = mp.player_id THEN 1 ELSE 0 END), 0) AS total_wins,
			COALESCE(SUM(CASE WHEN m.winner_id IS NOT NULL AND m.winner_id <> mp.player_id THEN 1 ELSE 0 END), 0) AS total_losses`).
		Joins("JOIN matches m ON m.id = mp.match_id").
		Where("mp.player_id = ? AND m.status IN ? AND m.mode = ?", playerID, finishedStatuses, string(mode)).
		Scan(&r).Error
	if err != nil {
		return PlayerStats{}, err
	}
	return statsOf(models.Players{TotalMatches: r.TotalMatches, TotalWins: r.TotalWins, TotalLosses: r.TotalLosses}), nil
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
