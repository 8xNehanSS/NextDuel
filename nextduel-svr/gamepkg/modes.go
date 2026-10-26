package gamepkg

import (
	"context"
	"nextduel-svr/questions"
	"nextduel-svr/utils"
	"time"
)

// Mode is one game NextDuel can host. Every mode shares the same duel
// structure, driven by Room: N rounds, each round offers the same choices to
// both players, the first answer shortens the opponent's clock, and then the
// mode decides who scored and who won the round. Match winner = most rounds won.
//
// To add a game: implement Mode, add its ID to utils.GameMode and register it
// in main.go (newModes). The frontend needs a matching entry in src/games.ts.
type Mode interface {
	Info() ModeInfo
	// Rounds returns n prompts. Choices must be non-empty; Answer may be ""
	// for modes without a single correct answer (e.g. rock-paper-scissors).
	Rounds(ctx context.Context, n int) ([]questions.Question, error)
	// Resolve scores one finished round. It runs under Room.mu.
	Resolve(q questions.Question, answers []RoundAnswer) RoundOutcome
}

type ModeInfo struct {
	ID utils.GameMode
	// RoundDuration caps the round clock for this mode. The configured
	// Settings.RoundDuration still applies if it is shorter. 0 = no cap.
	RoundDuration time.Duration
}

// RoundAnswer is one player's input for a round, in slot order.
type RoundAnswer struct {
	PlayerID string
	Answered bool
	Answer   string
	Elapsed  time.Duration // since the round started
	// Order is 1 for the first answer of the round, 2 for the second...
	// It is exact (assigned under Room.mu), unlike Elapsed, whose clock
	// resolution can make near-simultaneous answers look identical.
	Order int
}

type RoundOutcome struct {
	// Scored lists who earns +1 score this round.
	Scored map[string]bool
	// Winner of the round, or nil for nobody.
	Winner *string
	// CorrectAnswer is revealed to players; "" when there is none.
	CorrectAnswer string
	// Summary is a short human-readable explanation, e.g. "Paper covers Rock".
	Summary string
}

// Modes is the registry of playable games, keyed by ID.
type Modes map[utils.GameMode]Mode

func NewModes(modes ...Mode) Modes {
	m := make(Modes, len(modes))
	for _, mode := range modes {
		m[mode.Info().ID] = mode
	}
	return m
}

// ---------------------------------------------------------------------------
// Shared resolvers

// resolveExclusive: everyone correct scores; the round goes to a player only
// if they are the *only* one correct. Rewards knowledge (trivia).
func resolveExclusive(q questions.Question, answers []RoundAnswer) RoundOutcome {
	out := RoundOutcome{Scored: map[string]bool{}, CorrectAnswer: q.Answer}
	var correct []string
	for _, a := range answers {
		if a.Answered && a.Answer == q.Answer {
			out.Scored[a.PlayerID] = true
			correct = append(correct, a.PlayerID)
		}
	}
	switch len(correct) {
	case 0:
		out.Summary = "Nobody got it right."
	case 1:
		out.Winner = &correct[0]
		out.Summary = correct[0] + " was the only one to get it right."
	default:
		out.Summary = "Both players got it right."
	}
	return out
}

// resolveFastest: everyone correct scores; the round goes to whoever
// answered correctly first. Rewards speed (math, word games).
func resolveFastest(q questions.Question, answers []RoundAnswer) RoundOutcome {
	out := RoundOutcome{Scored: map[string]bool{}, CorrectAnswer: q.Answer}
	var best *RoundAnswer
	for i := range answers {
		a := &answers[i]
		if !a.Answered || a.Answer != q.Answer {
			continue
		}
		out.Scored[a.PlayerID] = true
		if best == nil || a.Order < best.Order {
			best = a
		}
	}
	if best == nil {
		out.Summary = "Nobody got it right."
		return out
	}
	id := best.PlayerID
	out.Winner = &id
	if len(out.Scored) > 1 {
		out.Summary = id + " was faster."
	} else {
		out.Summary = id + " was the only one to get it right."
	}
	return out
}
