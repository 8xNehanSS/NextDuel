// Package questions supplies trivia questions for duels.
//
// The default chain is OpenTDB first (fresh questions, cached into the local
// bank) and the local bank as a fallback. OpenTDB rate-limits to one request
// per ~5s per IP, so under concurrent load the bank does most of the work.
package questions

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
)

// Question is a single multiple-choice question. Choices always contains
// Answer and is already shuffled.
type Question struct {
	ID         string   `json:"id"`
	Text       string   `json:"question"`
	Answer     string   `json:"-"`
	Choices    []string `json:"choices"`
	Category   string   `json:"category"`
	Difficulty string   `json:"difficulty"`
}

// HasChoice reports whether answer is one of the offered choices.
func (q Question) HasChoice(answer string) bool {
	for _, c := range q.Choices {
		if c == answer {
			return true
		}
	}
	return false
}

// Provider returns n distinct questions or an error.
type Provider interface {
	Questions(ctx context.Context, n int) ([]Question, error)
}

var ErrNotEnoughQuestions = errors.New("not enough questions available")

func buildQuestion(id, text, answer string, incorrect []string, category, difficulty string) Question {
	choices := make([]string, 0, len(incorrect)+1)
	choices = append(choices, answer)
	choices = append(choices, incorrect...)
	rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })
	return Question{
		ID:         id,
		Text:       text,
		Answer:     answer,
		Choices:    choices,
		Category:   category,
		Difficulty: difficulty,
	}
}

// Fallback tries Primary and falls back to Secondary when Primary fails or
// returns too few questions. Successful primary results are handed to Cache.
type Fallback struct {
	Primary   Provider
	Secondary Provider
	Cache     func(ctx context.Context, qs []Question)
}

func (f *Fallback) Questions(ctx context.Context, n int) ([]Question, error) {
	if f.Primary != nil {
		qs, err := f.Primary.Questions(ctx, n)
		if err == nil && len(qs) >= n {
			if f.Cache != nil {
				f.Cache(ctx, qs)
			}
			return qs[:n], nil
		}
		slog.Warn("primary question provider failed, using fallback", "error", err, "got", len(qs), "want", n)
	}
	if f.Secondary == nil {
		return nil, ErrNotEnoughQuestions
	}
	return f.Secondary.Questions(ctx, n)
}
