// Package repository holds all SQL access outside of AutoMigrate.
package repository

import (
	"encoding/json"
	"errors"
	"log/slog"
	"nextduel-svr/gamepkg"
	"nextduel-svr/models"
	"nextduel-svr/questions"
	"nextduel-svr/utils"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MatchRecorder implements gamepkg.Recorder on top of GORM.
type MatchRecorder struct {
	db *gorm.DB

	mu        sync.Mutex
	playerIDs map[string]uint // username -> players.id cache
}

var _ gamepkg.Recorder = (*MatchRecorder)(nil)

func NewMatchRecorder(db *gorm.DB) *MatchRecorder {
	return &MatchRecorder{db: db, playerIDs: make(map[string]uint)}
}

func (m *MatchRecorder) StartMatch(roomID string, bestOf int, players []string) uint {
	ids, err := m.resolvePlayers(players)
	if err != nil {
		slog.Error("recorder: resolve players", "room", roomID, "error", err)
		return 0
	}
	var matchID uint
	err = m.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		match := models.Match{
			Mode:      string(utils.MatchModeClassic),
			Status:    string(utils.MatchStatusActive),
			BestOf:    bestOf,
			CreatedBy: ids[0],
			StartedAt: now,
		}
		if err := tx.Create(&match).Error; err != nil {
			return err
		}
		matchID = match.ID
		if err := logEvent(tx, matchID, nil, nil, utils.MatchEventCreated, map[string]any{"room_id": roomID}); err != nil {
			return err
		}
		for slot, id := range ids {
			if err := tx.Create(&models.MatchPlayer{MatchID: matchID, PlayerID: id, Slot: slot + 1, JoinedAt: now}).Error; err != nil {
				return err
			}
			pid := id
			if err := logEvent(tx, matchID, nil, &pid, utils.MatchEventPlayerJoin, map[string]any{"slot": slot + 1}); err != nil {
				return err
			}
		}
		return logEvent(tx, matchID, nil, nil, utils.MatchEventStarted, nil)
	})
	if err != nil {
		slog.Error("recorder: start match", "room", roomID, "error", err)
		return 0
	}
	return matchID
}

func (m *MatchRecorder) StartRound(matchID uint, roundNo int, q questions.Question) uint {
	if matchID == 0 {
		return 0
	}
	round := models.MatchRound{
		MatchID:   matchID,
		RoundNo:   roundNo,
		QuestID:   q.ID,
		Question:  q.Text,
		StartedAt: time.Now(),
	}
	if err := m.db.Create(&round).Error; err != nil {
		slog.Error("recorder: start round", "match", matchID, "round", roundNo, "error", err)
		return 0
	}
	rid := round.ID
	if err := logEvent(m.db, matchID, &rid, nil, utils.MatchEventRoundStart, map[string]any{"round": roundNo}); err != nil {
		slog.Warn("recorder: log round start", "error", err)
	}
	return round.ID
}

func (m *MatchRecorder) RecordAnswer(matchID, roundID uint, player, answer string, at time.Time) {
	if matchID == 0 || roundID == 0 {
		return
	}
	pid, err := m.resolvePlayer(player)
	if err != nil {
		slog.Error("recorder: resolve player", "player", player, "error", err)
		return
	}
	entry := models.RoundPlayer{RoundID: roundID, PlayerID: pid, Answer: answer, AnsweredAt: at}
	err = m.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "round_id"}, {Name: "player_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"answer", "answered_at"}),
	}).Create(&entry).Error
	if err != nil {
		slog.Error("recorder: record answer", "match", matchID, "error", err)
		return
	}
	if err := logEvent(m.db, matchID, &roundID, &pid, utils.MatchEventDuelAnswer, nil); err != nil {
		slog.Warn("recorder: log answer", "error", err)
	}
}

func (m *MatchRecorder) ResolveRound(matchID, roundID uint, roundNo int, winner *string, results []gamepkg.PlayerRoundResult) {
	if matchID == 0 || roundID == 0 {
		return
	}
	winnerID, err := m.optionalPlayer(winner)
	if err != nil {
		slog.Error("recorder: resolve winner", "match", matchID, "error", err)
		return
	}
	pids := make(map[string]uint, len(results))
	for _, res := range results {
		pid, err := m.resolvePlayer(res.PlayerID)
		if err != nil {
			slog.Error("recorder: resolve player", "player", res.PlayerID, "error", err)
			return
		}
		pids[res.PlayerID] = pid
	}

	err = m.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.MatchRound{}).Where("id = ?", roundID).
			Updates(map[string]any{"ended_at": time.Now(), "winner_id": winnerID}).Error; err != nil {
			return err
		}

		scores := make(map[string]int, len(results))
		for _, res := range results {
			pid := pids[res.PlayerID]
			isWinner := winner != nil && *winner == res.PlayerID
			entry := models.RoundPlayer{
				RoundID: roundID, PlayerID: pid, Answer: res.Answer,
				ScoreDelta: res.Delta, IsWinner: isWinner,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "round_id"}, {Name: "player_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"score_delta", "is_winner"}),
			}).Create(&entry).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.MatchPlayer{}).Where("match_id = ? AND player_id = ?", matchID, pid).
				Updates(map[string]any{"score": res.Score, "rounds_won": res.RoundsWon}).Error; err != nil {
				return err
			}
			rid := roundID
			if err := tx.Create(&models.ScoreEvent{
				MatchID: matchID, RoundID: &rid, PlayerID: pid,
				EventType: string(utils.ScoreEventRoundScore), Reason: string(utils.ScoreReasonRoundEnd),
				Delta: res.Delta, Total: res.Score,
			}).Error; err != nil {
				return err
			}
			scores[res.PlayerID] = res.Score
		}
		return logEvent(tx, matchID, &roundID, winnerID, utils.MatchEventRoundResult, map[string]any{"round": roundNo, "scores": scores})
	})
	if err != nil {
		slog.Error("recorder: resolve round", "match", matchID, "round", roundNo, "error", err)
	}
}

func (m *MatchRecorder) CompleteMatch(matchID uint, status utils.MatchStatus, winner *string, players []string, reason string) {
	if matchID == 0 {
		return
	}
	winnerID, err := m.optionalPlayer(winner)
	if err != nil {
		slog.Error("recorder: resolve winner", "match", matchID, "error", err)
		return
	}
	ids, err := m.resolvePlayers(players)
	if err != nil {
		slog.Error("recorder: resolve players", "match", matchID, "error", err)
		return
	}

	err = m.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&models.Match{}).Where("id = ?", matchID).
			Updates(map[string]any{"status": string(status), "winner_id": winnerID, "ended_at": now}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.MatchPlayer{}).Where("match_id = ?", matchID).
			Update("left_at", now).Error; err != nil {
			return err
		}
		for _, pid := range ids {
			updates := map[string]any{"total_matches": gorm.Expr("total_matches + 1")}
			if winnerID != nil {
				if pid == *winnerID {
					updates["total_wins"] = gorm.Expr("total_wins + 1")
				} else {
					updates["total_losses"] = gorm.Expr("total_losses + 1")
				}
			}
			if err := tx.Model(&models.Players{}).Where("id = ?", pid).Updates(updates).Error; err != nil {
				return err
			}
		}
		return logEvent(tx, matchID, nil, winnerID, utils.MatchEventDuelDone, map[string]any{"reason": reason})
	})
	if err != nil {
		slog.Error("recorder: complete match", "match", matchID, "error", err)
	}
}

func (m *MatchRecorder) AbortMatch(matchID uint, reason string) {
	if matchID == 0 {
		return
	}
	err := m.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Match{}).Where("id = ?", matchID).
			Updates(map[string]any{"status": string(utils.MatchStatusAborted), "ended_at": time.Now()}).Error; err != nil {
			return err
		}
		return logEvent(tx, matchID, nil, nil, utils.MatchEventDuelDone, map[string]any{"reason": reason})
	})
	if err != nil {
		slog.Error("recorder: abort match", "match", matchID, "error", err)
	}
}

// ---------------------------------------------------------------------------

func (m *MatchRecorder) optionalPlayer(name *string) (*uint, error) {
	if name == nil {
		return nil, nil
	}
	id, err := m.resolvePlayer(*name)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (m *MatchRecorder) resolvePlayers(names []string) ([]uint, error) {
	ids := make([]uint, len(names))
	for i, name := range names {
		id, err := m.resolvePlayer(name)
		if err != nil {
			return nil, err
		}
		ids[i] = id
	}
	return ids, nil
}

// resolvePlayer maps a username to its players row, creating it on first use.
// It runs outside any transaction: SQLite allows a single writer, so creating
// rows on m.db while a transaction is open would block on our own lock.
func (m *MatchRecorder) resolvePlayer(username string) (uint, error) {
	m.mu.Lock()
	id, ok := m.playerIDs[username]
	m.mu.Unlock()
	if ok {
		return id, nil
	}

	var player models.Players
	err := m.db.Where("username = ?", username).First(&player).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		player = models.Players{Username: username}
		err = m.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&player).Error
		if err == nil && player.ID == 0 {
			err = m.db.Where("username = ?", username).First(&player).Error
		}
	}
	if err != nil {
		return 0, err
	}

	m.mu.Lock()
	m.playerIDs[username] = player.ID
	m.mu.Unlock()
	return player.ID, nil
}

func logEvent(tx *gorm.DB, matchID uint, roundID, playerID *uint, kind utils.MatchEventType, payload any) error {
	var encoded string
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		encoded = string(b)
	}
	return tx.Create(&models.MatchEventLog{
		MatchID:    matchID,
		RoundID:    roundID,
		PlayerID:   playerID,
		EventType:  string(kind),
		Payload:    encoded,
		OccurredAt: time.Now(),
	}).Error
}
