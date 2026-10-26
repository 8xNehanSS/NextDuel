package gamepkg

import (
	"errors"
	"nextduel-svr/utils"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestQueuePairsPlayersIntoOneRoom(t *testing.T) {
	h := newHarness(t, nil)
	p1, c1 := h.connect("alice")
	p2, c2 := h.connect("bob")

	must(t, h.gm.JoinQueue(p1))
	if pos := c1.expect(t, utils.GameActionWaitingQueue).Data.(QueuePayload).Position; pos != 1 {
		t.Fatalf("queue position = %d, want 1", pos)
	}
	if err := h.gm.JoinQueue(p1); !errors.Is(err, ErrAlreadyQueued) {
		t.Fatalf("double queue err = %v, want ErrAlreadyQueued", err)
	}

	must(t, h.gm.JoinQueue(p2))
	state := c2.expect(t, utils.GameActionMatchFound).Data.(RoomState)
	c1.expect(t, utils.GameActionMatchFound)

	if r1, r2 := h.gm.RoomOf("alice"), h.gm.RoomOf("bob"); r1 == nil || r1 != r2 || r1.ID != state.RoomID {
		t.Fatalf("players not mapped to the same room: %v %v", r1, r2)
	}
	if len(state.Players) != 2 {
		t.Fatalf("room has %d players, want 2", len(state.Players))
	}
}

func TestLeaveQueue(t *testing.T) {
	h := newHarness(t, nil)
	p1, c1 := h.connect("alice")

	if err := h.gm.LeaveQueue(p1); !errors.Is(err, ErrNotQueued) {
		t.Fatalf("err = %v, want ErrNotQueued", err)
	}
	must(t, h.gm.JoinQueue(p1))
	must(t, h.gm.LeaveQueue(p1))
	c1.expect(t, utils.GameActionQueueLeft)
	if h.gm.Statistics().Queue != 0 {
		t.Fatal("queue not empty after leaving")
	}
}

func TestFullDuelProducesWinner(t *testing.T) {
	h := newHarness(t, nil)
	p1, p2, c1, c2 := h.startDuel(t)

	// Round 1: only alice is right, so she wins the round.
	must(t, h.gm.SubmitAnswer(p1, 1, "A"))
	must(t, h.gm.SubmitAnswer(p2, 1, "B"))
	res := c1.expect(t, utils.GameActionRoundResult).Data.(RoundResultPayload)
	if res.WinnerID == nil || *res.WinnerID != "alice" || res.CorrectAnswer != "A" {
		t.Fatalf("round 1 result = %+v", res)
	}

	// Round 2: both right, so nobody wins the round.
	c1.expect(t, utils.GameActionRoundStart)
	c2.expect(t, utils.GameActionRoundStart)
	must(t, h.gm.SubmitAnswer(p1, 2, "A"))
	must(t, h.gm.SubmitAnswer(p2, 2, "A"))
	res = c2.expect(t, utils.GameActionRoundResult).Data.(RoundResultPayload)
	if res.WinnerID != nil {
		t.Fatalf("round 2 should be a draw, got winner %q", *res.WinnerID)
	}

	over := c1.expect(t, utils.GameActionGameOver).Data.(GameOverPayload)
	if over.Reason != utils.GameOverCompleted || over.WinnerID == nil || *over.WinnerID != "alice" {
		t.Fatalf("game over = %+v", over)
	}
	scores := map[string]int{}
	for _, p := range over.Players {
		scores[p.ID] = p.Score
	}
	if scores["alice"] != 2 || scores["bob"] != 1 {
		t.Fatalf("final scores = %v, want alice=2 bob=1", scores)
	}

	eventually(t, func() bool { return h.gm.RoomOf("alice") == nil && h.gm.RoomOf("bob") == nil }, "room cleanup")
	comps := h.rec.completions()
	if len(comps) != 1 || comps[0].status != utils.MatchStatusCompleted || *comps[0].winner != "alice" {
		t.Fatalf("recorded completions = %+v", comps)
	}

	// Both players are free to queue again.
	must(t, h.gm.JoinQueue(p1))
}

func TestFirstAnswerTriggersFastDuel(t *testing.T) {
	h := newHarness(t, nil)
	p1, _, c1, c2 := h.startDuel(t)

	start := time.Now()
	must(t, h.gm.SubmitAnswer(p1, 1, "A"))
	c1.expect(t, utils.GameActionDuelAnswer)
	fast := c2.expect(t, utils.GameActionDuelFast).Data.(DuelFastPayload)
	if fast.PlayerID != "alice" || fast.Remaining > 1 {
		t.Fatalf("duel_fast = %+v", fast)
	}

	// Bob never answers; the round must end on the shortened clock, well
	// before the 2s round duration.
	c2.expect(t, utils.GameActionRoundResult)
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("round took %v after fast duel, want ~100ms", elapsed)
	}
}

func TestAnswerValidation(t *testing.T) {
	h := newHarness(t, nil)
	p1, _, _, _ := h.startDuel(t)

	cases := []struct {
		round  int
		answer string
		want   error
	}{
		{1, "Z", ErrInvalidAnswer},
		{2, "A", ErrWrongRound},
	}
	for _, tc := range cases {
		if err := h.gm.SubmitAnswer(p1, tc.round, tc.answer); !errors.Is(err, tc.want) {
			t.Errorf("SubmitAnswer(%d, %q) = %v, want %v", tc.round, tc.answer, err, tc.want)
		}
	}
	must(t, h.gm.SubmitAnswer(p1, 1, "B"))
	if err := h.gm.SubmitAnswer(p1, 1, "A"); !errors.Is(err, ErrAlreadyAnswered) {
		t.Errorf("second answer err = %v, want ErrAlreadyAnswered", err)
	}
}

func TestDisconnectMidMatchForfeits(t *testing.T) {
	h := newHarness(t, nil)
	_, _, c1, c2 := h.startDuel(t)

	h.gm.EndSession(c2)

	over := c1.expect(t, utils.GameActionGameOver).Data.(GameOverPayload)
	if over.Reason != utils.GameOverForfeit || over.WinnerID == nil || *over.WinnerID != "alice" {
		t.Fatalf("game over = %+v", over)
	}
	eventually(t, func() bool { return h.gm.RoomOf("alice") == nil }, "room cleanup")
	comps := h.rec.completions()
	if len(comps) != 1 || comps[0].status != utils.MatchStatusAbandoned || len(comps[0].players) != 2 {
		t.Fatalf("recorded completions = %+v", comps)
	}
}

func TestPrivateRoomFlow(t *testing.T) {
	h := newHarness(t, nil)
	host, hostConn := h.connect("alice")
	guest, guestConn := h.connect("bob")

	must(t, h.gm.CreateRoom(host))
	code := hostConn.expect(t, utils.GameActionRoomCreated).Data.(RoomState).RoomID
	if len(code) != utils.RoomCodeLength {
		t.Fatalf("room code %q has wrong length", code)
	}

	// Readying up alone must not start a solo game.
	must(t, h.gm.SetReady(host))
	hostConn.expectNone(t, utils.GameActionDuelStart, 50*time.Millisecond)

	if err := h.gm.JoinRoom(guest, "NOPE00"); !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("join unknown room err = %v", err)
	}
	must(t, h.gm.JoinRoom(guest, "  "+strings.ToLower(code)+" "))
	guestConn.expect(t, utils.GameActionRoomJoined)
	if n := len(hostConn.expect(t, utils.GameActionRoomState).Data.(RoomState).Players); n != 2 {
		t.Fatalf("host sees %d players, want 2", n)
	}

	must(t, h.gm.SetReady(guest))
	hostConn.expect(t, utils.GameActionDuelStart)
	guestConn.expect(t, utils.GameActionRoundStart)
}

func TestQueueRoomsAreNotJoinableByCode(t *testing.T) {
	h := newHarness(t, nil)
	p1, c1 := h.connect("alice")
	p2, _ := h.connect("bob")
	p3, _ := h.connect("carol")
	must(t, h.gm.JoinQueue(p1))
	must(t, h.gm.JoinQueue(p2))
	code := c1.expect(t, utils.GameActionMatchFound).Data.(RoomState).RoomID

	if err := h.gm.JoinRoom(p3, code); !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("err = %v, want ErrRoomNotFound", err)
	}
}

func TestLeavingWaitingRoom(t *testing.T) {
	h := newHarness(t, nil)
	host, hostConn := h.connect("alice")
	guest, _ := h.connect("bob")
	must(t, h.gm.CreateRoom(host))
	code := hostConn.expect(t, utils.GameActionRoomCreated).Data.(RoomState).RoomID
	must(t, h.gm.JoinRoom(guest, code))
	must(t, h.gm.SetReady(host))

	must(t, h.gm.LeaveRoom(guest))
	hostConn.expect(t, utils.GameActionPlayerLeft)
	state := hostConn.expect(t, utils.GameActionRoomState).Data.(RoomState)
	if len(state.Players) != 1 || state.Players[0].Ready {
		t.Fatalf("after guest left: %+v (host should be alone and un-readied)", state)
	}
	if h.gm.RoomOf("bob") != nil || h.gm.RoomOf("alice") == nil {
		t.Fatal("room mapping wrong after guest left")
	}

	must(t, h.gm.LeaveRoom(host))
	if h.gm.Statistics().Rooms != 0 {
		t.Fatal("empty room was not removed")
	}
}

func TestQuestionFailureAbortsRoom(t *testing.T) {
	h := newHarness(t, staticProvider{err: errors.New("upstream down")})
	p1, c1 := h.connect("alice")
	p2, _ := h.connect("bob")
	must(t, h.gm.JoinQueue(p1))
	must(t, h.gm.JoinQueue(p2))
	must(t, h.gm.SetReady(p1))
	must(t, h.gm.SetReady(p2))

	closed := c1.expect(t, utils.GameActionRoomDestroyed).Data.(RoomDestroyedPayload)
	if closed.Reason != utils.RoomCloseNoQuestions {
		t.Fatalf("reason = %q", closed.Reason)
	}
	eventually(t, func() bool { return h.gm.RoomOf("alice") == nil }, "room cleanup")
}

func TestSecondTabCannotActInRoom(t *testing.T) {
	h := newHarness(t, nil)
	p1, _, _, _ := h.startDuel(t)
	tab2, _ := h.connect("alice")

	if err := h.gm.SubmitAnswer(tab2, 1, "A"); !errors.Is(err, ErrNotInRoom) {
		t.Fatalf("second tab answer err = %v, want ErrNotInRoom", err)
	}
	// Closing the second tab must not forfeit the first tab's match.
	h.gm.EndSession(tab2.conn)
	must(t, h.gm.SubmitAnswer(p1, 1, "A"))
}

func TestShutdownAbortsRooms(t *testing.T) {
	h := newHarness(t, nil)
	_, _, c1, _ := h.startDuel(t)

	h.gm.Shutdown()
	if r := c1.expect(t, utils.GameActionRoomDestroyed).Data.(RoomDestroyedPayload); r.Reason != utils.RoomCloseServerShutdown {
		t.Fatalf("reason = %q", r.Reason)
	}
	p3, _ := h.connect("carol")
	if err := h.gm.JoinQueue(p3); !errors.Is(err, ErrShuttingDown) {
		t.Fatalf("join during shutdown err = %v", err)
	}
}

// Hammers the manager from many goroutines; run with -race where available.
func TestConcurrentMatchmaking(t *testing.T) {
	h := newHarness(t, nil)
	const n = 50
	var wg sync.WaitGroup
	conns := make([]*fakeConn, n)
	for i := 0; i < n; i++ {
		p, c := h.connect(string(rune('a'+i%26)) + string(rune('A'+i/26)))
		conns[i] = c
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = h.gm.JoinQueue(p)
			_ = h.gm.SetReady(p)
		}()
	}
	wg.Wait()
	for _, c := range conns {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.gm.EndSession(c)
		}()
	}
	wg.Wait()
	eventually(t, func() bool {
		s := h.gm.Statistics()
		return s.Rooms == 0 && s.Queue == 0 && s.Players == 0
	}, "all rooms cleaned up")
}
