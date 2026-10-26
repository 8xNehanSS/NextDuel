package gamepkg

import (
	"context"
	"errors"
	"nextduel-svr/questions"
	"nextduel-svr/utils"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// answered builds an answer; order is derived from elapsed (seconds).
func answered(id, answer string, elapsed time.Duration) RoundAnswer {
	return RoundAnswer{PlayerID: id, Answered: true, Answer: answer, Elapsed: elapsed, Order: int(elapsed/time.Second) + 1}
}

func winnerOf(o RoundOutcome) string {
	if o.Winner == nil {
		return ""
	}
	return *o.Winner
}

func TestResolvers(t *testing.T) {
	q := questions.Question{Answer: "A", Choices: []string{"A", "B"}}
	silent := RoundAnswer{PlayerID: "bob"}

	cases := []struct {
		name    string
		resolve func(questions.Question, []RoundAnswer) RoundOutcome
		answers []RoundAnswer
		winner  string
		scored  int
	}{
		{"exclusive: only one right", resolveExclusive, []RoundAnswer{answered("alice", "A", 5*time.Second), answered("bob", "B", time.Second)}, "alice", 1},
		{"exclusive: both right is a draw", resolveExclusive, []RoundAnswer{answered("alice", "A", 5*time.Second), answered("bob", "A", time.Second)}, "", 2},
		{"exclusive: nobody right", resolveExclusive, []RoundAnswer{answered("alice", "B", 0), silent}, "", 0},
		{"fastest: both right, faster wins", resolveFastest, []RoundAnswer{answered("alice", "A", 5*time.Second), answered("bob", "A", time.Second)}, "bob", 2},
		{"fastest: fast but wrong loses", resolveFastest, []RoundAnswer{answered("alice", "A", 5*time.Second), answered("bob", "B", time.Second)}, "alice", 1},
		{"fastest: no answers", resolveFastest, []RoundAnswer{{PlayerID: "alice"}, silent}, "", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := tc.resolve(q, tc.answers)
			if got := winnerOf(out); got != tc.winner {
				t.Errorf("winner = %q, want %q", got, tc.winner)
			}
			if len(out.Scored) != tc.scored {
				t.Errorf("scored = %v, want %d players", out.Scored, tc.scored)
			}
			if out.CorrectAnswer != "A" || out.Summary == "" {
				t.Errorf("reveal = %q, summary = %q", out.CorrectAnswer, out.Summary)
			}
		})
	}
}

func TestRPSResolve(t *testing.T) {
	m := NewRPSMode()
	cases := []struct {
		a, b   string // "" = did not pick
		winner string
	}{
		{rpsRock, rpsScissors, "alice"},
		{rpsScissors, rpsRock, "bob"},
		{rpsPaper, rpsRock, "alice"},
		{rpsScissors, rpsPaper, "alice"},
		{rpsPaper, rpsScissors, "bob"},
		{rpsRock, rpsRock, ""},
		{rpsPaper, "", "alice"},
		{"", rpsPaper, "bob"},
		{"", "", ""},
	}
	pick := func(id, choice string) RoundAnswer {
		if choice == "" {
			return RoundAnswer{PlayerID: id}
		}
		return answered(id, choice, 0)
	}
	for _, tc := range cases {
		out := m.Resolve(questions.Question{}, []RoundAnswer{pick("alice", tc.a), pick("bob", tc.b)})
		if got := winnerOf(out); got != tc.winner {
			t.Errorf("%q vs %q: winner = %q, want %q (%s)", tc.a, tc.b, got, tc.winner, out.Summary)
		}
		// In RPS the round winner is the only one who scores.
		if (tc.winner == "") != (len(out.Scored) == 0) || (tc.winner != "" && !out.Scored[tc.winner]) {
			t.Errorf("%q vs %q: scored = %v", tc.a, tc.b, out.Scored)
		}
		if out.CorrectAnswer != "" {
			t.Errorf("RPS should reveal no correct answer, got %q", out.CorrectAnswer)
		}
	}
}

func TestMathProblemsAreCorrect(t *testing.T) {
	m := NewMathMode()
	for iter := 0; iter < 200; iter++ {
		qs, err := m.Rounds(context.Background(), 5)
		if err != nil {
			t.Fatal(err)
		}
		for _, q := range qs {
			assertChoices(t, q, 4)
			got, err := evalMath(strings.TrimSuffix(q.Text, " = ?"))
			if err != nil {
				t.Fatalf("%q: %v", q.Text, err)
			}
			if strconv.Itoa(got) != q.Answer {
				t.Fatalf("%q: answer %s, actual %d", q.Text, q.Answer, got)
			}
		}
	}
}

func TestScramblePuzzlesAreFair(t *testing.T) {
	m := NewScrambleMode()
	seen := map[string]bool{}
	for _, w := range m.words {
		if seen[w] {
			t.Fatalf("duplicate word %q", w)
		}
		seen[w] = true
	}
	for iter := 0; iter < 100; iter++ {
		qs, err := m.Rounds(context.Background(), 5)
		if err != nil {
			t.Fatal(err)
		}
		for _, q := range qs {
			assertChoices(t, q, 4)
			prompt := strings.ReplaceAll(q.Text, " ", "")
			if prompt == q.Answer {
				t.Fatalf("prompt is not scrambled: %q", q.Text)
			}
			for _, c := range append([]string{prompt}, q.Choices...) {
				if sortedLetters(c) != sortedLetters(q.Answer) {
					t.Fatalf("%q is not an arrangement of %q", c, q.Answer)
				}
			}
		}
	}
	if _, err := m.Rounds(context.Background(), len(m.words)+1); !errors.Is(err, questions.ErrNotEnoughQuestions) {
		t.Fatalf("too many rounds err = %v", err)
	}
}

func TestModeCapsRoundDuration(t *testing.T) {
	long := testSettings()
	long.RoundDuration = time.Minute
	if d := newRoom("X", false, long, NewRPSMode(), nil, nil).roundDuration(); d != 10*time.Second {
		t.Errorf("rps with 60s config: %v, want the mode's 10s cap", d)
	}
	if d := newRoom("X", false, long, NewTriviaMode(staticProvider{}), nil, nil).roundDuration(); d != time.Minute {
		t.Errorf("trivia: %v, want configured 60s", d)
	}
	short := testSettings() // 2s, shorter than any cap
	if d := newRoom("X", false, short, NewMathMode(), nil, nil).roundDuration(); d != short.RoundDuration {
		t.Errorf("math with 2s config: %v, want 2s", d)
	}
}

// --- room-level flows per mode ----------------------------------------------

func TestQueuesAreSeparatePerMode(t *testing.T) {
	h := newHarness(t, nil)
	alice, ca := h.connect("alice")
	bob, cb := h.connect("bob")
	carol, _ := h.connect("carol")

	must(t, h.gm.JoinQueue(alice, utils.GameModeTrivia))
	must(t, h.gm.JoinQueue(bob, utils.GameModeRPS))
	if q := cb.expect(t, utils.GameActionWaitingQueue).Data.(QueuePayload); q.Mode != utils.GameModeRPS || q.Position != 1 {
		t.Fatalf("bob queue = %+v", q)
	}
	ca.expectNone(t, utils.GameActionMatchFound, 30*time.Millisecond)

	must(t, h.gm.JoinQueue(carol, utils.GameModeRPS))
	if state := cb.expect(t, utils.GameActionMatchFound).Data.(RoomState); state.Mode != utils.GameModeRPS {
		t.Fatalf("matched room mode = %q", state.Mode)
	}
	if s := h.gm.Statistics(); s.Queues[utils.GameModeTrivia] != 1 || s.Queues[utils.GameModeRPS] != 0 {
		t.Fatalf("queues = %v", s.Queues)
	}
	if err := h.gm.JoinQueue(alice, utils.GameModeMath); !errors.Is(err, ErrAlreadyQueued) {
		t.Fatalf("queueing for a second mode: %v, want ErrAlreadyQueued", err)
	}
}

func TestUnknownModeRejected(t *testing.T) {
	h := newHarness(t, nil)
	p, _ := h.connect("alice")
	if err := h.gm.JoinQueue(p, "chess"); !errors.Is(err, ErrUnknownMode) {
		t.Fatalf("JoinQueue err = %v", err)
	}
	if err := h.gm.CreateRoom(p, "chess"); !errors.Is(err, ErrUnknownMode) {
		t.Fatalf("CreateRoom err = %v", err)
	}
	must(t, h.gm.CreateRoom(p, "")) // empty = trivia, for older clients
}

func startModeDuel(t *testing.T, h *harness, mode utils.GameMode) (p1, p2 *Player, c1, c2 *fakeConn) {
	t.Helper()
	p1, c1 = h.connect("alice")
	p2, c2 = h.connect("bob")
	must(t, h.gm.CreateRoom(p1, mode))
	code := c1.expect(t, utils.GameActionRoomCreated).Data.(RoomState).RoomID
	must(t, h.gm.JoinRoom(p2, code))
	must(t, h.gm.SetReady(p1))
	must(t, h.gm.SetReady(p2))
	c1.expect(t, utils.GameActionRoundStart)
	c2.expect(t, utils.GameActionRoundStart)
	return
}

func TestRPSDuel(t *testing.T) {
	h := newHarness(t, nil)
	p1, p2, c1, c2 := startModeDuel(t, h, utils.GameModeRPS)

	must(t, h.gm.SubmitAnswer(p1, 1, rpsRock))
	// The opponent learns that alice picked, but not what.
	if fast := c2.expect(t, utils.GameActionDuelFast).Data.(DuelFastPayload); fast.PlayerID != "alice" {
		t.Fatalf("duel_fast = %+v", fast)
	}
	must(t, h.gm.SubmitAnswer(p2, 1, rpsScissors))
	res := c1.expect(t, utils.GameActionRoundResult).Data.(RoundResultPayload)
	if winnerOf(RoundOutcome{Winner: res.WinnerID}) != "alice" || res.Summary != "Rock crushes Scissors." {
		t.Fatalf("round 1 = %+v", res)
	}

	c1.expect(t, utils.GameActionRoundStart)
	c2.expect(t, utils.GameActionRoundStart)
	must(t, h.gm.SubmitAnswer(p1, 2, rpsPaper))
	must(t, h.gm.SubmitAnswer(p2, 2, rpsPaper))
	if res := c2.expect(t, utils.GameActionRoundResult).Data.(RoundResultPayload); res.WinnerID != nil {
		t.Fatalf("tie round has winner %q", *res.WinnerID)
	}

	over := c1.expect(t, utils.GameActionGameOver).Data.(GameOverPayload)
	if over.WinnerID == nil || *over.WinnerID != "alice" {
		t.Fatalf("game over = %+v", over)
	}
}

func TestFastestCorrectWinsMathRound(t *testing.T) {
	h := newHarness(t, nil)
	p1, p2, c1, _ := startModeDuel(t, h, utils.GameModeMath)
	answer := currentAnswer(t, h, "alice")

	// Both correct; bob answers first, so bob takes the round.
	must(t, h.gm.SubmitAnswer(p2, 1, answer))
	must(t, h.gm.SubmitAnswer(p1, 1, answer))
	res := c1.expect(t, utils.GameActionRoundResult).Data.(RoundResultPayload)
	if res.WinnerID == nil || *res.WinnerID != "bob" || res.CorrectAnswer != answer {
		t.Fatalf("round result = %+v", res)
	}
	for _, r := range res.Results {
		if !r.Correct || r.Score != 1 {
			t.Fatalf("both players should score: %+v", res.Results)
		}
	}
}

// --- helpers ----------------------------------------------------------------

func currentAnswer(t *testing.T, h *harness, playerID string) string {
	t.Helper()
	room := h.gm.RoomOf(playerID)
	if room == nil {
		t.Fatal("player not in a room")
	}
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.questions[room.round-1].Answer
}

func assertChoices(t *testing.T, q questions.Question, n int) {
	t.Helper()
	if len(q.Choices) != n {
		t.Fatalf("%q: %d choices, want %d: %v", q.Text, len(q.Choices), n, q.Choices)
	}
	unique := map[string]bool{}
	for _, c := range q.Choices {
		unique[c] = true
	}
	if len(unique) != n || !q.HasChoice(q.Answer) {
		t.Fatalf("%q: choices %v must be distinct and include %q", q.Text, q.Choices, q.Answer)
	}
}

func evalMath(expr string) (int, error) {
	parts := strings.Fields(expr)
	if len(parts) != 3 {
		return 0, errors.New("unexpected expression")
	}
	a, err1 := strconv.Atoi(parts[0])
	b, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil {
		return 0, errors.New("bad operand")
	}
	switch parts[1] {
	case "+":
		return a + b, nil
	case "−":
		return a - b, nil
	case "×":
		return a * b, nil
	case "÷":
		if b == 0 || a%b != 0 {
			return 0, errors.New("inexact division")
		}
		return a / b, nil
	}
	return 0, errors.New("unknown operator " + parts[1])
}

func sortedLetters(s string) string {
	r := []rune(s)
	sort.Slice(r, func(i, j int) bool { return r[i] < r[j] })
	return string(r)
}
