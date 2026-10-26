package gamepkg

import (
	"context"
	"fmt"
	"math/rand/v2"
	"nextduel-svr/questions"
	"nextduel-svr/utils"
	"strconv"
	"time"
)

// MathMode generates mental-arithmetic problems that get harder each round.
// Fastest correct answer wins the round.
type MathMode struct{}

func NewMathMode() *MathMode { return &MathMode{} }

func (m *MathMode) Info() ModeInfo {
	return ModeInfo{ID: utils.GameModeMath, RoundDuration: 20 * time.Second}
}

func (m *MathMode) Rounds(_ context.Context, n int) ([]questions.Question, error) {
	out := make([]questions.Question, n)
	for i := range out {
		out[i] = mathProblem(i, n)
	}
	return out, nil
}

func (m *MathMode) Resolve(q questions.Question, answers []RoundAnswer) RoundOutcome {
	return resolveFastest(q, answers)
}

// mathProblem builds problem i of n; difficulty ramps from 0 to 1.
func mathProblem(i, n int) questions.Question {
	level := 0.0
	if n > 1 {
		level = float64(i) / float64(n-1)
	}
	difficulty := "easy"
	switch {
	case level > 0.66:
		difficulty = "hard"
	case level > 0.33:
		difficulty = "medium"
	}

	// Operand ranges grow with level.
	small := 10 + int(level*40) // 10..50
	big := 20 + int(level*180)  // 20..200

	var text string
	var answer int
	switch rand.IntN(4) {
	case 0:
		a, b := rand.IntN(big)+1, rand.IntN(big)+1
		text, answer = fmt.Sprintf("%d + %d", a, b), a+b
	case 1:
		a, b := rand.IntN(big)+1, rand.IntN(big)+1
		if b > a {
			a, b = b, a
		}
		text, answer = fmt.Sprintf("%d − %d", a, b), a-b
	case 2:
		a, b := rand.IntN(small/2)+2, rand.IntN(small/2)+2
		text, answer = fmt.Sprintf("%d × %d", a, b), a*b
	default:
		b, q := rand.IntN(small/3)+2, rand.IntN(small/2)+2
		text, answer = fmt.Sprintf("%d ÷ %d", b*q, b), q
	}

	choices := []string{strconv.Itoa(answer)}
	seen := map[int]bool{answer: true}
	spread := max(3, answer/10)
	for len(choices) < 4 {
		d := answer + rand.IntN(2*spread+1) - spread
		if d < 0 || seen[d] {
			continue
		}
		seen[d] = true
		choices = append(choices, strconv.Itoa(d))
	}
	rand.Shuffle(len(choices), func(a, b int) { choices[a], choices[b] = choices[b], choices[a] })

	return questions.Question{
		ID:         fmt.Sprintf("math-%d-%s", i+1, text),
		Text:       text + " = ?",
		Answer:     strconv.Itoa(answer),
		Choices:    choices,
		Category:   "Mental math",
		Difficulty: difficulty,
	}
}
