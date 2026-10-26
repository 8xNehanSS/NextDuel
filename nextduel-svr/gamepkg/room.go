package gamepkg

import (
	"context"
	"log/slog"
	"math"
	"nextduel-svr/questions"
	"nextduel-svr/utils"
	"sync"
	"time"
)

const maxPlayers = 2

type Settings struct {
	Rounds               int
	RoundDuration        time.Duration
	FastDuelDuration     time.Duration
	IntermissionDuration time.Duration
	// TickInterval is how often timers are broadcast. Tests shrink it.
	TickInterval    time.Duration
	QuestionTimeout time.Duration
}

func DefaultSettings() Settings {
	return Settings{
		Rounds:               5,
		RoundDuration:        60 * time.Second,
		FastDuelDuration:     10 * time.Second,
		IntermissionDuration: 5 * time.Second,
		TickInterval:         time.Second,
		QuestionTimeout:      8 * time.Second,
	}
}

type roomPlayer struct {
	player    *Player
	ready     bool
	score     int
	roundsWon int
	answered  bool
	answer    string
}

// Room is a single duel between two players.
//
// Concurrency: every field below mu is guarded by mu. The match itself runs in
// one goroutine (run) started when both players are ready; websocket handlers
// call AddPlayer/SetReady/SubmitAnswer/Leave concurrently with it.
//
// Lock ordering: GameManager.mu may be held while calling AddPlayer, SetReady,
// SubmitAnswer or State. Leave and Abort may close the room, which calls back
// into the manager, so they must be called without GameManager.mu held.
type Room struct {
	ID      string
	Private bool

	settings Settings
	provider questions.Provider
	recorder Recorder
	onClosed func(roomID string, playerIDs []string)

	wake chan struct{}
	stop chan struct{}

	mu          sync.Mutex
	status      utils.RoomStatus
	closed      bool
	players     []*roomPlayer
	questions   []questions.Question
	round       int
	accepting   bool
	deadline    time.Time
	fastApplied bool
	matchID     uint
	roundID     uint
}

func newRoom(id string, private bool, settings Settings, provider questions.Provider, recorder Recorder, onClosed func(string, []string)) *Room {
	if recorder == nil {
		recorder = noopRecorder{}
	}
	return &Room{
		ID:       id,
		Private:  private,
		settings: settings,
		provider: provider,
		recorder: recorder,
		onClosed: onClosed,
		wake:     make(chan struct{}, 1),
		stop:     make(chan struct{}),
		status:   utils.RoomStatusWaiting,
	}
}

// ---------------------------------------------------------------------------
// Commands (called from websocket handlers)

func (r *Room) AddPlayer(p *Player, announce utils.GameAction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch {
	case r.closed:
		return ErrRoomClosed
	case r.status != utils.RoomStatusWaiting:
		return ErrMatchStarted
	case r.findByIDLocked(p.ID) != nil:
		return ErrAlreadyInRoom
	case len(r.players) >= maxPlayers:
		return ErrRoomFull
	}

	r.players = append(r.players, &roomPlayer{player: p})
	state := r.stateLocked()
	p.Send(announce, state)
	r.broadcastExceptLocked(p, utils.GameActionRoomState, state)
	return nil
}

func (r *Room) SetReady(p *Player) error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return ErrRoomClosed
	}
	rp := r.findLocked(p)
	if rp == nil {
		r.mu.Unlock()
		return ErrNotInRoom
	}
	if r.status != utils.RoomStatusWaiting {
		r.mu.Unlock()
		return ErrMatchStarted
	}

	rp.ready = true
	start := len(r.players) == maxPlayers
	for _, other := range r.players {
		start = start && other.ready
	}
	if start {
		r.status = utils.RoomStatusStarting
	}
	r.broadcastLocked(utils.GameActionRoomState, r.stateLocked())
	r.mu.Unlock()

	if start {
		go r.run()
	}
	return nil
}

func (r *Room) SubmitAnswer(p *Player, round int, answer string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rp := r.findLocked(p)
	switch {
	case rp == nil:
		return ErrNotInRoom
	case r.closed || r.status != utils.RoomStatusActive || !r.accepting:
		return ErrMatchNotActive
	case round != r.round:
		return ErrWrongRound
	case rp.answered:
		return ErrAlreadyAnswered
	case !r.questions[r.round-1].HasChoice(answer):
		return ErrInvalidAnswer
	}

	now := time.Now()
	rp.answered = true
	rp.answer = answer
	r.recorder.RecordAnswer(r.matchID, r.roundID, p.ID, answer, now)
	p.Send(utils.GameActionDuelAnswer, AnswerAckPayload{Round: round, Answer: answer})

	// The first answer shortens the clock for whoever is still thinking.
	if !r.fastApplied && !r.allAnsweredLocked() {
		r.fastApplied = true
		if fast := now.Add(r.settings.FastDuelDuration); fast.Before(r.deadline) {
			r.deadline = fast
		}
		r.broadcastLocked(utils.GameActionDuelFast, DuelFastPayload{
			PlayerID:  p.ID,
			Remaining: secondsUntil(now, r.deadline),
		})
	}

	r.signal()
	return nil
}

// Leave removes p. Before the match starts the room survives if someone is
// still in it; once the match is underway leaving forfeits it.
func (r *Room) Leave(p *Player) error {
	r.mu.Lock()
	rp := r.findLocked(p)
	if rp == nil || r.closed {
		r.mu.Unlock()
		return ErrNotInRoom
	}

	shouldClose := false
	switch r.status {
	case utils.RoomStatusWaiting:
		r.removeLocked(rp)
		if len(r.players) == 0 {
			shouldClose = true
			break
		}
		for _, other := range r.players {
			other.ready = false
		}
		r.broadcastLocked(utils.GameActionPlayerLeft, PlayerLeftPayload{PlayerID: p.ID})
		r.broadcastLocked(utils.GameActionRoomState, r.stateLocked())

	case utils.RoomStatusStarting:
		// Nothing has been recorded yet, so there is no result to award.
		r.removeLocked(rp)
		r.broadcastLocked(utils.GameActionRoomDestroyed, RoomDestroyedPayload{RoomID: r.ID, Reason: utils.RoomCloseOpponentLeft})
		shouldClose = true

	case utils.RoomStatusActive:
		allIDs := r.playerIDsLocked()
		r.removeLocked(rp)
		var winner *string
		if len(r.players) > 0 {
			id := r.players[0].player.ID
			winner = &id
		}
		r.recorder.CompleteMatch(r.matchID, utils.MatchStatusAbandoned, winner, allIDs, string(utils.ScoreReasonDisconnect))
		standings := append(r.standingsLocked(), PlayerState{ID: rp.player.ID, Score: rp.score, RoundsWon: rp.roundsWon})
		r.broadcastLocked(utils.GameActionGameOver, GameOverPayload{
			Reason:   utils.GameOverForfeit,
			WinnerID: winner,
			Players:  standings,
		})
		shouldClose = true
	}

	var ids []string
	if shouldClose {
		ids = r.closeLocked()
	}
	r.mu.Unlock()

	if shouldClose {
		r.notifyClosed(ids)
	}
	return nil
}

// Abort closes the room without a result.
func (r *Room) Abort(reason utils.RoomCloseReason) {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	if r.matchID != 0 {
		r.recorder.AbortMatch(r.matchID, string(reason))
	}
	r.broadcastLocked(utils.GameActionRoomDestroyed, RoomDestroyedPayload{RoomID: r.ID, Reason: reason})
	ids := r.closeLocked()
	r.mu.Unlock()
	r.notifyClosed(ids)
}

func (r *Room) State() RoomState {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stateLocked()
}

func (r *Room) Status() utils.RoomStatus {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.status
}

// ---------------------------------------------------------------------------
// Match loop

func (r *Room) run() {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("room loop panicked", "room", r.ID, "panic", rec)
			r.Abort(utils.RoomCloseNoQuestions)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), r.settings.QuestionTimeout)
	qs, err := r.provider.Questions(ctx, r.settings.Rounds)
	cancel()
	if err != nil || len(qs) < r.settings.Rounds {
		slog.Error("could not load questions", "room", r.ID, "error", err, "got", len(qs))
		r.Abort(utils.RoomCloseNoQuestions)
		return
	}

	r.mu.Lock()
	if r.closed || r.status != utils.RoomStatusStarting {
		r.mu.Unlock()
		return
	}
	r.questions = qs[:r.settings.Rounds]
	r.status = utils.RoomStatusActive
	r.matchID = r.recorder.StartMatch(r.ID, len(r.questions), r.playerIDsLocked())
	r.broadcastLocked(utils.GameActionDuelStart, DuelStartPayload{MaxRounds: len(r.questions)})
	r.mu.Unlock()

	for i := range r.questions {
		if !r.playRound(i + 1) {
			return
		}
		if i < len(r.questions)-1 && !r.intermission(i+2) {
			return
		}
	}
	r.finish()
}

func (r *Room) playRound(n int) bool {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return false
	}
	q := r.questions[n-1]
	r.round = n
	for _, rp := range r.players {
		rp.answered = false
		rp.answer = ""
	}
	r.accepting = true
	r.fastApplied = false
	r.deadline = time.Now().Add(r.settings.RoundDuration)
	r.roundID = r.recorder.StartRound(r.matchID, n, q)
	r.broadcastLocked(utils.GameActionRoundStart, RoundStartPayload{
		Round:      n,
		MaxRounds:  len(r.questions),
		Question:   q.Text,
		Category:   q.Category,
		Difficulty: q.Difficulty,
		Choices:    q.Choices,
		Duration:   int(math.Ceil(r.settings.RoundDuration.Seconds())),
	})
	r.mu.Unlock()

	ticker := time.NewTicker(r.settings.TickInterval)
	defer ticker.Stop()
	for {
		select {
		case <-r.stop:
			return false
		case <-r.wake:
		case <-ticker.C:
		}

		r.mu.Lock()
		if r.closed {
			r.mu.Unlock()
			return false
		}
		now := time.Now()
		if r.allAnsweredLocked() || !now.Before(r.deadline) {
			r.resolveRoundLocked()
			r.mu.Unlock()
			return true
		}
		r.broadcastLocked(utils.GameActionTimerUpdate, TimerPayload{Remaining: secondsUntil(now, r.deadline)})
		r.mu.Unlock()
	}
}

func (r *Room) intermission(nextRound int) bool {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return false
	}
	end := time.Now().Add(r.settings.IntermissionDuration)
	r.broadcastLocked(utils.GameActionIntermission, IntermissionPayload{
		Remaining: secondsUntil(time.Now(), end),
		NextRound: nextRound,
	})
	r.mu.Unlock()

	ticker := time.NewTicker(r.settings.TickInterval)
	defer ticker.Stop()
	for {
		select {
		case <-r.stop:
			return false
		case <-ticker.C:
		}

		r.mu.Lock()
		if r.closed {
			r.mu.Unlock()
			return false
		}
		now := time.Now()
		if !now.Before(end) {
			r.mu.Unlock()
			return true
		}
		r.broadcastLocked(utils.GameActionIntermissionTimerUpdate, TimerPayload{Remaining: secondsUntil(now, end)})
		r.mu.Unlock()
	}
}

func (r *Room) resolveRoundLocked() {
	r.accepting = false
	q := r.questions[r.round-1]

	var correct []*roomPlayer
	for _, rp := range r.players {
		if rp.answered && rp.answer == q.Answer {
			rp.score++
			correct = append(correct, rp)
		}
	}

	// A round is won by being the only player to answer correctly.
	var winner *string
	if len(correct) == 1 {
		correct[0].roundsWon++
		id := correct[0].player.ID
		winner = &id
	}

	results := make([]PlayerRoundResult, 0, len(r.players))
	for _, rp := range r.players {
		isCorrect := rp.answered && rp.answer == q.Answer
		delta := 0
		if isCorrect {
			delta = 1
		}
		results = append(results, PlayerRoundResult{
			PlayerID:  rp.player.ID,
			Answer:    rp.answer,
			Correct:   isCorrect,
			Delta:     delta,
			Score:     rp.score,
			RoundsWon: rp.roundsWon,
		})
	}

	r.recorder.ResolveRound(r.matchID, r.roundID, r.round, winner, results)
	r.broadcastLocked(utils.GameActionRoundResult, RoundResultPayload{
		Round:         r.round,
		CorrectAnswer: q.Answer,
		WinnerID:      winner,
		Results:       results,
	})
}

func (r *Room) finish() {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}

	var winner *string
	if len(r.players) == maxPlayers {
		a, b := r.players[0], r.players[1]
		switch {
		case a.roundsWon > b.roundsWon:
			winner = &a.player.ID
		case b.roundsWon > a.roundsWon:
			winner = &b.player.ID
		}
	}

	r.recorder.CompleteMatch(r.matchID, utils.MatchStatusCompleted, winner, r.playerIDsLocked(), string(utils.GameOverCompleted))
	r.broadcastLocked(utils.GameActionGameOver, GameOverPayload{
		Reason:   utils.GameOverCompleted,
		WinnerID: winner,
		Players:  r.standingsLocked(),
	})
	ids := r.closeLocked()
	r.mu.Unlock()
	r.notifyClosed(ids)
}

// ---------------------------------------------------------------------------
// Helpers (all *Locked helpers require r.mu)

func (r *Room) signal() {
	select {
	case r.wake <- struct{}{}:
	default:
	}
}

func (r *Room) closeLocked() []string {
	if r.closed {
		return nil
	}
	r.closed = true
	r.status = utils.RoomStatusCompleted
	r.accepting = false
	close(r.stop)
	return r.playerIDsLocked()
}

func (r *Room) notifyClosed(ids []string) {
	if r.onClosed != nil {
		r.onClosed(r.ID, ids)
	}
}

func (r *Room) findLocked(p *Player) *roomPlayer {
	for _, rp := range r.players {
		if rp.player == p {
			return rp
		}
	}
	return nil
}

func (r *Room) findByIDLocked(id string) *roomPlayer {
	for _, rp := range r.players {
		if rp.player.ID == id {
			return rp
		}
	}
	return nil
}

func (r *Room) removeLocked(target *roomPlayer) {
	for i, rp := range r.players {
		if rp == target {
			r.players = append(r.players[:i], r.players[i+1:]...)
			return
		}
	}
}

func (r *Room) allAnsweredLocked() bool {
	for _, rp := range r.players {
		if !rp.answered {
			return false
		}
	}
	return len(r.players) > 0
}

func (r *Room) playerIDsLocked() []string {
	ids := make([]string, len(r.players))
	for i, rp := range r.players {
		ids[i] = rp.player.ID
	}
	return ids
}

func (r *Room) standingsLocked() []PlayerState {
	out := make([]PlayerState, len(r.players))
	for i, rp := range r.players {
		out[i] = PlayerState{ID: rp.player.ID, Ready: rp.ready, Score: rp.score, RoundsWon: rp.roundsWon}
	}
	return out
}

func (r *Room) stateLocked() RoomState {
	maxRounds := len(r.questions)
	if maxRounds == 0 {
		maxRounds = r.settings.Rounds
	}
	return RoomState{
		RoomID:    r.ID,
		Private:   r.Private,
		Status:    r.status,
		Round:     r.round,
		MaxRounds: maxRounds,
		Players:   r.standingsLocked(),
	}
}

func (r *Room) broadcastLocked(action utils.GameAction, data any) {
	for _, rp := range r.players {
		rp.player.Send(action, data)
	}
}

func (r *Room) broadcastExceptLocked(skip *Player, action utils.GameAction, data any) {
	for _, rp := range r.players {
		if rp.player != skip {
			rp.player.Send(action, data)
		}
	}
}

func secondsUntil(now, t time.Time) int {
	d := t.Sub(now)
	if d <= 0 {
		return 0
	}
	return int(math.Ceil(d.Seconds()))
}
