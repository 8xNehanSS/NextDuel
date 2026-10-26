package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"nextduel-svr/config"
	"nextduel-svr/database"
	"nextduel-svr/models"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// TestEndToEnd boots the full HTTP stack and plays a duel over real
// websockets: register -> login -> queue -> ready -> answer -> history.
func TestEndToEnd(t *testing.T) {
	cfg := config.Config{
		DBPath:         filepath.Join(t.TempDir(), "e2e.db"),
		OpenTDBEnabled: false,
		Game: config.GameConfig{
			Rounds:               2,
			RoundDuration:        3 * time.Second,
			FastDuelDuration:     300 * time.Millisecond,
			IntermissionDuration: 50 * time.Millisecond,
		},
	}
	db, err := database.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
	handler, game := newApp(cfg, db)
	srv := httptest.NewServer(handler)
	t.Cleanup(func() {
		game.Shutdown()
		srv.Close()
	})

	alice, bob := newClient(t, srv), newClient(t, srv)

	// --- auth -------------------------------------------------------------
	alice.expect(http.MethodGet, "/api/user", nil, http.StatusUnauthorized)
	alice.expect(http.MethodPost, "/api/register", map[string]string{"username": "x", "email": "bad", "password": "1"}, http.StatusBadRequest)
	alice.expect(http.MethodPost, "/api/register", creds("alice"), http.StatusCreated)
	bob.expect(http.MethodPost, "/api/register", creds("bob"), http.StatusCreated)
	bob.expect(http.MethodPost, "/api/register", creds("ALICE"), http.StatusConflict)
	alice.expect(http.MethodPost, "/api/login", map[string]string{"username": "alice", "password": "wrong-password"}, http.StatusUnauthorized)
	alice.expect(http.MethodPost, "/api/login", creds("alice"), http.StatusOK)
	bob.expect(http.MethodPost, "/api/login", creds("bob"), http.StatusOK)

	var me struct{ Username string }
	alice.decode(alice.expect(http.MethodGet, "/api/user", nil, http.StatusOK), &me)
	if me.Username != "alice" {
		t.Fatalf("GET /api/user = %q", me.Username)
	}
	alice.expect(http.MethodGet, "/api/health", nil, http.StatusOK)
	alice.expect(http.MethodGet, "/api/does-not-exist", nil, http.StatusNotFound)

	// --- websocket origin check -------------------------------------------
	if _, resp, err := alice.dial("https://evil.example"); err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-origin websocket was not rejected: err=%v resp=%v", err, resp)
	}

	// --- duel ---------------------------------------------------------------
	wa, _, err := alice.dial("")
	if err != nil {
		t.Fatal(err)
	}
	defer wa.Close()
	wb, _, err := bob.dial("")
	if err != nil {
		t.Fatal(err)
	}
	defer wb.Close()

	send(t, wa, "join_queue", nil)
	readUntil(t, wa, "waiting_queue")
	send(t, wb, "join_queue", nil)
	readUntil(t, wa, "match_found")
	readUntil(t, wb, "match_found")

	send(t, wa, "player_ready", nil)
	send(t, wb, "player_ready", nil)

	for round := 1; round <= cfg.Game.Rounds; round++ {
		var q struct {
			Round    int      `json:"round"`
			Question string   `json:"question"`
			Choices  []string `json:"choices"`
		}
		json.Unmarshal(readUntil(t, wa, "round_start"), &q)
		readUntil(t, wb, "round_start")
		correct := answerFor(t, db, q.Question)
		wrong := q.Choices[0]
		if wrong == correct {
			wrong = q.Choices[1]
		}

		send(t, wa, "duel_answer", map[string]any{"round": q.Round, "answer": correct})
		readUntil(t, wa, "duel_answer")
		readUntil(t, wb, "duel_fast")
		send(t, wb, "duel_answer", map[string]any{"round": q.Round, "answer": wrong})

		var res struct {
			WinnerID *string `json:"winnerId"`
		}
		json.Unmarshal(readUntil(t, wb, "round_result"), &res)
		if res.WinnerID == nil || *res.WinnerID != "alice" {
			t.Fatalf("round %d winner = %v, want alice", round, res.WinnerID)
		}
	}

	var over struct {
		Reason   string  `json:"reason"`
		WinnerID *string `json:"winnerId"`
	}
	json.Unmarshal(readUntil(t, wa, "game_over"), &over)
	if over.Reason != "completed" || over.WinnerID == nil || *over.WinnerID != "alice" {
		t.Fatalf("game_over = %+v", over)
	}

	// --- persistence --------------------------------------------------------
	var history struct {
		Total   int
		Stats   struct{ Wins int }
		Matches []struct {
			Result   string
			Opponent string
		}
	}
	alice.decode(alice.expect(http.MethodGet, "/api/history/duels", nil, http.StatusOK), &history)
	if history.Total != 1 || history.Stats.Wins != 1 || history.Matches[0].Result != "win" || history.Matches[0].Opponent != "bob" {
		t.Fatalf("history = %+v", history)
	}
	var board []struct {
		Rank     int
		Username string
	}
	bob.decode(bob.expect(http.MethodGet, "/api/leaderboard", nil, http.StatusOK), &board)
	if len(board) != 2 || board[0].Username != "alice" {
		t.Fatalf("leaderboard = %+v", board)
	}

	// --- another game: rock-paper-scissors in a private room ------------------
	send(t, wa, "create_room", map[string]string{"mode": "rps"})
	var room struct {
		RoomID string `json:"roomId"`
		Mode   string `json:"mode"`
	}
	json.Unmarshal(readUntil(t, wa, "room_created"), &room)
	if room.Mode != "rps" {
		t.Fatalf("room mode = %q, want rps", room.Mode)
	}
	send(t, wb, "join_room", map[string]string{"roomId": room.RoomID})
	readUntil(t, wb, "room_joined")
	send(t, wa, "player_ready", nil)
	send(t, wb, "player_ready", nil)

	picks := [][2]string{{"Rock", "Scissors"}, {"Paper", "Paper"}}
	for i, pick := range picks {
		readUntil(t, wa, "round_start")
		readUntil(t, wb, "round_start")
		send(t, wa, "duel_answer", map[string]any{"round": i + 1, "answer": pick[0]})
		send(t, wb, "duel_answer", map[string]any{"round": i + 1, "answer": pick[1]})
		readUntil(t, wa, "round_result")
	}
	json.Unmarshal(readUntil(t, wb, "game_over"), &over)
	if over.WinnerID == nil || *over.WinnerID != "alice" {
		t.Fatalf("rps game_over = %+v", over)
	}

	alice.decode(alice.expect(http.MethodGet, "/api/history/duels?mode=rps", nil, http.StatusOK), &history)
	if history.Total != 1 || history.Matches[0].Result != "win" {
		t.Fatalf("rps history = %+v", history)
	}
	bob.decode(bob.expect(http.MethodGet, "/api/leaderboard?mode=rps", nil, http.StatusOK), &board)
	if len(board) != 2 || board[0].Username != "alice" {
		t.Fatalf("rps leaderboard = %+v", board)
	}

	// --- rate limiting --------------------------------------------------------
	limited := false
	for i := 0; i < 10 && !limited; i++ {
		limited = bob.do(http.MethodPost, "/api/login", creds("bob")).StatusCode == http.StatusTooManyRequests
	}
	if !limited {
		t.Fatal("login was never rate limited")
	}
}

// ---------------------------------------------------------------------------

type client struct {
	t   *testing.T
	srv *httptest.Server
	*http.Client
}

func newClient(t *testing.T, srv *httptest.Server) *client {
	jar, _ := cookiejar.New(nil)
	return &client{t: t, srv: srv, Client: &http.Client{Jar: jar}}
}

func creds(name string) map[string]string {
	return map[string]string{"username": name, "email": strings.ToLower(name) + "@example.com", "password": "correct-horse"}
}

func (c *client) do(method, path string, body any) *http.Response {
	c.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, c.srv.URL+path, &buf)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	return resp
}

func (c *client) expect(method, path string, body any, status int) *http.Response {
	c.t.Helper()
	resp := c.do(method, path, body)
	if resp.StatusCode != status {
		var b bytes.Buffer
		b.ReadFrom(resp.Body)
		c.t.Fatalf("%s %s = %d, want %d: %s", method, path, resp.StatusCode, status, b.String())
	}
	return resp
}

// decode unwraps nex-server's {success, data} envelope.
func (c *client) decode(resp *http.Response, dest any) {
	c.t.Helper()
	defer resp.Body.Close()
	var env struct{ Data json.RawMessage }
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		c.t.Fatal(err)
	}
	if err := json.Unmarshal(env.Data, dest); err != nil {
		c.t.Fatal(err)
	}
}

func (c *client) dial(origin string) (*websocket.Conn, *http.Response, error) {
	u, _ := url.Parse(c.srv.URL)
	header := http.Header{}
	for _, ck := range c.Jar.Cookies(u) {
		header.Add("Cookie", ck.String())
	}
	if origin != "" {
		header.Set("Origin", origin)
	}
	return websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(c.srv.URL, "http")+"/api/ws", header)
}

func send(t *testing.T, conn *websocket.Conn, kind string, data any) {
	t.Helper()
	if err := conn.WriteJSON(map[string]any{"type": kind, "data": data}); err != nil {
		t.Fatal(err)
	}
}

func readUntil(t *testing.T, conn *websocket.Conn, kind string) json.RawMessage {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		var msg struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("waiting for %q: %v", kind, err)
		}
		if msg.Type == "error" {
			t.Fatalf("server error while waiting for %q: %s", kind, msg.Data)
		}
		if msg.Type == kind {
			return msg.Data
		}
	}
}

func answerFor(t *testing.T, db *gorm.DB, question string) string {
	t.Helper()
	var q models.Quest
	if err := db.Where("question = ?", question).First(&q).Error; err != nil {
		t.Fatalf("question %q not in bank: %v", question, err)
	}
	return q.Answer
}
