package gamepkg

import (
	"context"
	"errors"
	"fmt"
	"nextduel-svr/questions"
	"nextduel-svr/utils"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const waitTimeout = 2 * time.Second

type fakeConn struct {
	msgs   chan Message
	closed atomic.Bool
}

func newFakeConn() *fakeConn { return &fakeConn{msgs: make(chan Message, 1024)} }

func (c *fakeConn) WriteJSON(v any) error {
	if c.closed.Load() {
		return errors.New("closed")
	}
	c.msgs <- v.(Message)
	return nil
}
func (c *fakeConn) SetWriteDeadline(time.Time) error { return nil }
func (c *fakeConn) Close() error                     { c.closed.Store(true); return nil }

// expect skips messages until one of the given type arrives.
func (c *fakeConn) expect(t *testing.T, action utils.GameAction) Message {
	t.Helper()
	deadline := time.After(waitTimeout)
	for {
		select {
		case m := <-c.msgs:
			if m.Type == action {
				return m
			}
		case <-deadline:
			t.Fatalf("timed out waiting for %q", action)
			return Message{}
		}
	}
}

// expectNone asserts no message of the given type arrives within d.
func (c *fakeConn) expectNone(t *testing.T, action utils.GameAction, d time.Duration) {
	t.Helper()
	deadline := time.After(d)
	for {
		select {
		case m := <-c.msgs:
			if m.Type == action {
				t.Fatalf("unexpected %q: %+v", action, m.Data)
			}
		case <-deadline:
			return
		}
	}
}

// staticProvider returns questions whose correct answer is always "A".
type staticProvider struct{ err error }

func (s staticProvider) Questions(_ context.Context, n int) ([]questions.Question, error) {
	if s.err != nil {
		return nil, s.err
	}
	out := make([]questions.Question, n)
	for i := range out {
		out[i] = questions.Question{
			ID:      fmt.Sprintf("q%d", i+1),
			Text:    fmt.Sprintf("Question %d?", i+1),
			Answer:  "A",
			Choices: []string{"A", "B", "C", "D"},
		}
	}
	return out, nil
}

type completion struct {
	status  utils.MatchStatus
	winner  *string
	players []string
}

type recordingRecorder struct {
	noopRecorder
	mu        sync.Mutex
	completed []completion
	aborted   int
}

func (r *recordingRecorder) StartMatch(string, utils.GameMode, int, []string) uint { return 1 }
func (r *recordingRecorder) CompleteMatch(_ uint, status utils.MatchStatus, winner *string, players []string, _ string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.completed = append(r.completed, completion{status, winner, players})
}
func (r *recordingRecorder) AbortMatch(uint, string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.aborted++
}
func (r *recordingRecorder) completions() []completion {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]completion(nil), r.completed...)
}

func testSettings() Settings {
	return Settings{
		Rounds:               2,
		RoundDuration:        2 * time.Second,
		FastDuelDuration:     100 * time.Millisecond,
		IntermissionDuration: 30 * time.Millisecond,
		TickInterval:         10 * time.Millisecond,
		QuestionTimeout:      time.Second,
	}
}

type harness struct {
	gm  *GameManager
	rec *recordingRecorder
}

func newHarness(t *testing.T, provider questions.Provider) *harness {
	t.Helper()
	if provider == nil {
		provider = staticProvider{}
	}
	rec := &recordingRecorder{}
	return &harness{gm: NewGameManager(testSettings(), NewModes(NewTriviaMode(provider), NewMathMode(), NewScrambleMode(), NewRPSMode()), rec), rec: rec}
}

func (h *harness) connect(id string) (*Player, *fakeConn) {
	conn := newFakeConn()
	return h.gm.Session(id, conn), conn
}

// matched queues two players and readies them up, returning once round 1
// has started for both.
func (h *harness) startDuel(t *testing.T) (p1, p2 *Player, c1, c2 *fakeConn) {
	t.Helper()
	p1, c1 = h.connect("alice")
	p2, c2 = h.connect("bob")
	must(t, h.gm.JoinQueue(p1, utils.GameModeTrivia))
	must(t, h.gm.JoinQueue(p2, utils.GameModeTrivia))
	c1.expect(t, utils.GameActionMatchFound)
	c2.expect(t, utils.GameActionMatchFound)
	must(t, h.gm.SetReady(p1))
	must(t, h.gm.SetReady(p2))
	c1.expect(t, utils.GameActionRoundStart)
	c2.expect(t, utils.GameActionRoundStart)
	return
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func eventually(t *testing.T, cond func() bool, what string) {
	t.Helper()
	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for: %s", what)
}
