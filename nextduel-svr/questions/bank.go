package questions

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"nextduel-svr/models"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:embed seed.json
var seedJSON []byte

// Bank serves questions from the local quest table.
type Bank struct {
	db *gorm.DB
}

func NewBank(db *gorm.DB) *Bank {
	return &Bank{db: db}
}

// HashID derives a stable ID from question text.
func HashID(text string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(text))))
	return hex.EncodeToString(sum[:20])
}

func (b *Bank) Questions(ctx context.Context, n int) ([]Question, error) {
	var rows []models.Quest
	if err := b.db.WithContext(ctx).Order("RANDOM()").Limit(n).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("question bank: %w", err)
	}
	if len(rows) < n {
		return nil, fmt.Errorf("%w: bank has %d, need %d", ErrNotEnoughQuestions, len(rows), n)
	}

	out := make([]Question, 0, len(rows))
	for _, row := range rows {
		var incorrect []string
		if err := json.Unmarshal([]byte(row.IncorrectAnswers), &incorrect); err != nil || len(incorrect) == 0 {
			slog.Warn("question bank: skipping malformed row", "id", row.ID)
			continue
		}
		out = append(out, buildQuestion(row.ID, row.Question, row.Answer, incorrect, row.Category, row.Difficulty))
	}
	if len(out) < n {
		return nil, ErrNotEnoughQuestions
	}
	return out, nil
}

// Store inserts questions, ignoring ones already present.
func (b *Bank) Store(ctx context.Context, qs []Question, source string) error {
	if len(qs) == 0 {
		return nil
	}
	rows := make([]models.Quest, 0, len(qs))
	for _, q := range qs {
		incorrect := make([]string, 0, len(q.Choices)-1)
		for _, c := range q.Choices {
			if c != q.Answer {
				incorrect = append(incorrect, c)
			}
		}
		encoded, err := json.Marshal(incorrect)
		if err != nil {
			return err
		}
		id := q.ID
		if id == "" {
			id = HashID(q.Text)
		}
		rows = append(rows, models.Quest{
			ID:               id,
			Question:         q.Text,
			Answer:           q.Answer,
			IncorrectAnswers: string(encoded),
			Difficulty:       q.Difficulty,
			Category:         q.Category,
			Source:           source,
		})
	}
	return b.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

// CacheFunc adapts Store for use as Fallback.Cache.
func (b *Bank) CacheFunc(source string) func(context.Context, []Question) {
	return func(ctx context.Context, qs []Question) {
		if err := b.Store(ctx, qs, source); err != nil {
			slog.Warn("question bank: failed to cache questions", "error", err)
		}
	}
}

type seedQuestion struct {
	Question   string   `json:"question"`
	Answer     string   `json:"answer"`
	Incorrect  []string `json:"incorrect"`
	Category   string   `json:"category"`
	Difficulty string   `json:"difficulty"`
}

// Seed loads the embedded starter questions. It is idempotent.
func (b *Bank) Seed(ctx context.Context) error {
	var seeds []seedQuestion
	if err := json.Unmarshal(seedJSON, &seeds); err != nil {
		return fmt.Errorf("parse seed questions: %w", err)
	}
	qs := make([]Question, 0, len(seeds))
	for _, s := range seeds {
		qs = append(qs, buildQuestion(HashID(s.Question), s.Question, s.Answer, s.Incorrect, s.Category, s.Difficulty))
	}
	return b.Store(ctx, qs, "seed")
}
