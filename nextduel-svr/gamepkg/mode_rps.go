package gamepkg

import (
	"context"
	"fmt"
	"nextduel-svr/questions"
	"nextduel-svr/utils"
	"time"
)

const (
	rpsRock     = "Rock"
	rpsPaper    = "Paper"
	rpsScissors = "Scissors"
)

// rpsBeats[x] is what x defeats, and how.
var rpsBeats = map[string]struct{ victim, verb string }{
	rpsRock:     {rpsScissors, "crushes"},
	rpsPaper:    {rpsRock, "covers"},
	rpsScissors: {rpsPaper, "cut"},
}

// RPSMode is rock-paper-scissors. Picks stay hidden until the round resolves
// (opponents only learn *that* you picked, via duel_fast). Not picking loses
// to any pick.
type RPSMode struct{}

func NewRPSMode() *RPSMode { return &RPSMode{} }

func (m *RPSMode) Info() ModeInfo {
	return ModeInfo{ID: utils.GameModeRPS, RoundDuration: 10 * time.Second}
}

func (m *RPSMode) Rounds(_ context.Context, n int) ([]questions.Question, error) {
	out := make([]questions.Question, n)
	for i := range out {
		out[i] = questions.Question{
			ID:       fmt.Sprintf("rps-%d", i+1),
			Text:     "Rock, paper or scissors?",
			Choices:  []string{rpsRock, rpsPaper, rpsScissors},
			Category: "Rock Paper Scissors",
		}
	}
	return out, nil
}

func (m *RPSMode) Resolve(_ questions.Question, answers []RoundAnswer) RoundOutcome {
	out := RoundOutcome{Scored: map[string]bool{}}
	if len(answers) != 2 {
		out.Summary = "Round void."
		return out
	}
	a, b := answers[0], answers[1]
	win := func(w RoundAnswer, summary string) RoundOutcome {
		id := w.PlayerID
		out.Winner = &id
		out.Scored[id] = true
		out.Summary = summary
		return out
	}

	switch {
	case !a.Answered && !b.Answered:
		out.Summary = "Nobody picked."
	case !b.Answered:
		return win(a, b.PlayerID+" didn't pick in time.")
	case !a.Answered:
		return win(b, a.PlayerID+" didn't pick in time.")
	case a.Answer == b.Answer:
		out.Summary = "Both picked " + a.Answer + ". It's a tie."
	case rpsBeats[a.Answer].victim == b.Answer:
		return win(a, a.Answer+" "+rpsBeats[a.Answer].verb+" "+b.Answer+".")
	default:
		return win(b, b.Answer+" "+rpsBeats[b.Answer].verb+" "+a.Answer+".")
	}
	return out
}
