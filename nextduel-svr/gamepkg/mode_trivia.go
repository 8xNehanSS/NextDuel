package gamepkg

import (
	"context"
	"nextduel-svr/questions"
	"nextduel-svr/utils"
)

// TriviaMode serves multiple-choice general-knowledge questions.
type TriviaMode struct {
	provider questions.Provider
}

func NewTriviaMode(provider questions.Provider) *TriviaMode {
	return &TriviaMode{provider: provider}
}

func (m *TriviaMode) Info() ModeInfo { return ModeInfo{ID: utils.GameModeTrivia} }

func (m *TriviaMode) Rounds(ctx context.Context, n int) ([]questions.Question, error) {
	return m.provider.Questions(ctx, n)
}

func (m *TriviaMode) Resolve(q questions.Question, answers []RoundAnswer) RoundOutcome {
	return resolveExclusive(q, answers)
}
