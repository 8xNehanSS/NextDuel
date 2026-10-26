package gamepkg

import (
	"context"
	"fmt"
	"math/rand/v2"
	"nextduel-svr/questions"
	"nextduel-svr/utils"
	"strings"
	"time"
)

// ScrambleMode shows a word with its letters shuffled. Every choice is an
// arrangement of the same letters; only one is a real word. Fastest correct
// answer wins the round.
type ScrambleMode struct {
	words []string
}

func NewScrambleMode() *ScrambleMode { return &ScrambleMode{words: scrambleWords} }

func (m *ScrambleMode) Info() ModeInfo {
	return ModeInfo{ID: utils.GameModeScramble, RoundDuration: 20 * time.Second}
}

func (m *ScrambleMode) Rounds(_ context.Context, n int) ([]questions.Question, error) {
	if n > len(m.words) {
		return nil, questions.ErrNotEnoughQuestions
	}
	picks := rand.Perm(len(m.words))[:n]
	out := make([]questions.Question, n)
	for i, idx := range picks {
		out[i] = scramblePuzzle(i, strings.ToUpper(m.words[idx]))
	}
	return out, nil
}

func (m *ScrambleMode) Resolve(q questions.Question, answers []RoundAnswer) RoundOutcome {
	return resolveFastest(q, answers)
}

func scramblePuzzle(i int, word string) questions.Question {
	letters := []rune(word)
	shuffled := shuffleUntilDifferent(letters, word)

	// Decoys are near-misses: the real word with two letters swapped, so the
	// player has to actually recognise the word rather than spot the odd one.
	choices := []string{word}
	seen := map[string]bool{word: true, shuffled: true}
	for attempts := 0; len(choices) < 4 && attempts < 200; attempts++ {
		d := swapTwo(letters)
		if !seen[d] {
			seen[d] = true
			choices = append(choices, d)
		}
	}
	// Very short or repetitive words may not yield three distinct swaps.
	for attempts := 0; len(choices) < 4 && attempts < 200; attempts++ {
		d := shuffleUntilDifferent(letters, word)
		if !seen[d] {
			seen[d] = true
			choices = append(choices, d)
		}
	}
	rand.Shuffle(len(choices), func(a, b int) { choices[a], choices[b] = choices[b], choices[a] })

	difficulty := "easy"
	switch {
	case len(letters) >= 8:
		difficulty = "hard"
	case len(letters) >= 6:
		difficulty = "medium"
	}
	return questions.Question{
		ID:         fmt.Sprintf("scramble-%d-%s", i+1, word),
		Text:       spaced(shuffled),
		Answer:     word,
		Choices:    choices,
		Category:   "Unscramble the word",
		Difficulty: difficulty,
	}
}

func shuffleUntilDifferent(letters []rune, word string) string {
	out := make([]rune, len(letters))
	for attempt := 0; attempt < 50; attempt++ {
		copy(out, letters)
		rand.Shuffle(len(out), func(a, b int) { out[a], out[b] = out[b], out[a] })
		if string(out) != word {
			return string(out)
		}
	}
	return string(out)
}

func swapTwo(letters []rune) string {
	out := append([]rune(nil), letters...)
	a := rand.IntN(len(out))
	b := rand.IntN(len(out))
	out[a], out[b] = out[b], out[a]
	return string(out)
}

func spaced(s string) string {
	return strings.Join(strings.Split(s, ""), " ")
}

// Common words of 5-8 letters with few duplicate letters, so near-miss
// decoys are clearly not words.
var scrambleWords = []string{
	"planet", "garden", "bright", "castle", "dragon", "forest", "guitar",
	"harbor", "island", "jungle", "kitten", "marble", "number", "orange",
	"pencil", "rocket", "sunset", "tunnel", "violet", "window", "yellow", "anchor",
	"basket", "candle", "dinner", "engine", "falcon", "galaxy", "hunter", "insect",
	"jacket", "knight", "ladder", "magnet", "napkin", "oyster", "pirate", "quartz",
	"rabbit", "spider", "turtle", "unicorn", "valley", "walrus", "zipper", "breeze",
	"cherry", "desert", "eclipse", "fossil", "glacier", "horizon", "iceberg", "journey",
	"kingdom", "lantern", "mystery", "network", "octopus", "penguin", "quantum", "rainbow",
	"sparrow", "thunder", "volcano", "whisper", "captain", "diamond", "crystal", "blanket",
	"chimney", "cabinet", "festival", "mountain", "elephant", "dinosaur", "keyboard", "treasure",
	"sandwich", "painting", "football", "airplane", "umbrella", "triangle", "compass", "pumpkin",
	"teacher", "wizard", "bridge", "cloud", "flame", "grape", "house", "juice",
	"light", "mango", "night", "piano", "queen", "river", "stone", "tiger",
	"water", "zebra", "brick", "chair", "dance", "eagle", "frost", "ghost",
}
