package repository

import (
	"nextduel-svr/database"
	"nextduel-svr/gamepkg"
	"nextduel-svr/models"
	"nextduel-svr/questions"
	"nextduel-svr/utils"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := gorm.Open(sqlite.Open(path+"?_pragma=busy_timeout(5000)"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
	return db
}

// playMatch records a trivia match through the recorder, exactly as a Room would.
func playMatch(rec *MatchRecorder, a, b string, winner *string, status utils.MatchStatus) uint {
	return playModeMatch(rec, utils.GameModeTrivia, a, b, winner, status)
}

func playModeMatch(rec *MatchRecorder, mode utils.GameMode, a, b string, winner *string, status utils.MatchStatus) uint {
	id := rec.StartMatch("ROOM01", mode, 2, []string{a, b})
	q := questions.Question{ID: "q1", Text: "Q?", Answer: "A", Choices: []string{"A", "B"}}
	round := rec.StartRound(id, 1, q)
	rec.RecordAnswer(id, round, a, "A", time.Now())
	rec.RecordAnswer(id, round, b, "B", time.Now())
	rec.ResolveRound(id, round, 1, &a, []gamepkg.PlayerRoundResult{
		{PlayerID: a, Answer: "A", Correct: true, Delta: 1, Score: 1, RoundsWon: 1},
		{PlayerID: b, Answer: "B", Delta: 0, Score: 0},
	})
	rec.CompleteMatch(id, status, winner, []string{a, b}, "test")
	return id
}

func TestRecorderAndHistory(t *testing.T) {
	db := openTestDB(t)
	rec := NewMatchRecorder(db)
	alice, bob := "alice", "bob"

	m1 := playMatch(rec, alice, bob, &alice, utils.MatchStatusCompleted)
	if m1 == 0 {
		t.Fatal("StartMatch returned 0")
	}
	playMatch(rec, alice, bob, nil, utils.MatchStatusCompleted)                                       // draw
	playMatch(rec, alice, bob, &bob, utils.MatchStatusAbandoned)                                      // alice forfeited
	rec.AbortMatch(rec.StartMatch("ROOM02", utils.GameModeTrivia, 2, []string{alice, bob}), "server") // aborted: not in stats

	var match models.Match
	if err := db.First(&match, m1).Error; err != nil {
		t.Fatal(err)
	}
	if match.Status != string(utils.MatchStatusCompleted) || match.WinnerID == nil {
		t.Fatalf("match row = %+v", match)
	}
	var rp []models.RoundPlayer
	db.Where("round_id IN (SELECT id FROM match_rounds WHERE match_id = ?)", m1).Find(&rp)
	if len(rp) != 2 {
		t.Fatalf("round_players rows = %d, want 2 (answer + resolve must upsert)", len(rp))
	}

	h, err := History(db, alice, "", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	// The aborted match has no result and must not appear.
	if h.Total != 3 || len(h.Matches) != 3 {
		t.Fatalf("history total=%d len=%d, want 3", h.Total, len(h.Matches))
	}
	if h.Stats.TotalMatches != 3 || h.Stats.Wins != 1 || h.Stats.Losses != 1 || h.Stats.Draws != 1 {
		t.Fatalf("stats = %+v", h.Stats)
	}
	results := map[string]int{}
	for _, m := range h.Matches {
		results[m.Result]++
		if m.Opponent != bob {
			t.Fatalf("opponent = %q, want bob", m.Opponent)
		}
	}
	if results["win"] != 1 || results["loss"] != 1 || results["draw"] != 1 {
		t.Fatalf("results = %v", results)
	}

	page, err := History(db, alice, "", 1, 1)
	if err != nil || len(page.Matches) != 1 || page.Total != 3 {
		t.Fatalf("pagination: %+v, %v", page, err)
	}
}

func TestHistoryForNewPlayerIsEmpty(t *testing.T) {
	db := openTestDB(t)
	h, err := History(db, "nobody", "", 10, 0)
	if err != nil {
		t.Fatalf("History for unknown user returned error: %v", err)
	}
	if h.Total != 0 || h.Matches == nil {
		t.Fatalf("want empty non-nil history, got %+v", h)
	}
}

func TestLeaderboardOrdering(t *testing.T) {
	db := openTestDB(t)
	rec := NewMatchRecorder(db)
	a, b, c := "alice", "bob", "carol"

	playMatch(rec, a, b, &a, utils.MatchStatusCompleted)
	playMatch(rec, a, c, &a, utils.MatchStatusCompleted)
	playMatch(rec, b, c, &b, utils.MatchStatusCompleted)
	playMatch(rec, b, c, &c, utils.MatchStatusCompleted)

	board, err := Leaderboard(db, "", 10)
	if err != nil {
		t.Fatal(err)
	}
	var order []string
	for _, e := range board {
		order = append(order, e.Username)
	}
	// alice 2/2, bob 1/3 (more matches, same wins as carol 1/3 -> name order)
	want := []string{"alice", "bob", "carol"}
	for i := range want {
		if i >= len(order) || order[i] != want[i] {
			t.Fatalf("leaderboard order = %v, want %v", order, want)
		}
	}
	if board[0].Rank != 1 || board[0].WinRate != 1 {
		t.Fatalf("top entry = %+v", board[0])
	}
}

func TestPerModeHistoryAndLeaderboard(t *testing.T) {
	db := openTestDB(t)
	rec := NewMatchRecorder(db)
	a, b := "alice", "bob"

	playModeMatch(rec, utils.GameModeTrivia, a, b, &a, utils.MatchStatusCompleted)
	playModeMatch(rec, utils.GameModeRPS, a, b, &b, utils.MatchStatusCompleted)
	playModeMatch(rec, utils.GameModeRPS, a, b, &b, utils.MatchStatusCompleted)
	playModeMatch(rec, utils.GameModeMath, a, b, nil, utils.MatchStatusCompleted)

	h, err := History(db, a, utils.GameModeRPS, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if h.Total != 2 || h.Stats.TotalMatches != 2 || h.Stats.Losses != 2 || h.Stats.Wins != 0 {
		t.Fatalf("rps history for alice = total %d, stats %+v", h.Total, h.Stats)
	}
	for _, m := range h.Matches {
		if m.Mode != string(utils.GameModeRPS) {
			t.Fatalf("filtered history contains %q", m.Mode)
		}
	}
	if all, _ := History(db, a, "", 10, 0); all.Total != 4 || all.Stats.TotalMatches != 4 || all.Stats.Draws != 1 {
		t.Fatalf("unfiltered history = total %d, stats %+v", all.Total, all.Stats)
	}
	if none, err := History(db, a, utils.GameModeScramble, 10, 0); err != nil || none.Total != 0 || none.Stats.TotalMatches != 0 {
		t.Fatalf("unplayed mode history = %+v, %v", none, err)
	}

	board, err := Leaderboard(db, utils.GameModeRPS, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(board) != 2 || board[0].Username != b || board[0].Wins != 2 || board[1].Losses != 2 {
		t.Fatalf("rps leaderboard = %+v", board)
	}
	trivia, _ := Leaderboard(db, utils.GameModeTrivia, 10)
	if len(trivia) != 2 || trivia[0].Username != a {
		t.Fatalf("trivia leaderboard = %+v", trivia)
	}
	overall, _ := Leaderboard(db, "", 10)
	if len(overall) != 2 || overall[0].Username != b || overall[0].TotalMatches != 4 {
		t.Fatalf("overall leaderboard = %+v", overall)
	}
}

func TestMigrateRenamesLegacyClassicMode(t *testing.T) {
	db := openTestDB(t)
	legacy := models.Match{Mode: string(utils.GameModeLegacyClassic), Status: string(utils.MatchStatusCompleted)}
	if err := db.Create(&legacy).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	var got models.Match
	db.First(&got, legacy.ID)
	if got.Mode != string(utils.GameModeTrivia) {
		t.Fatalf("legacy match mode = %q, want trivia", got.Mode)
	}
}
