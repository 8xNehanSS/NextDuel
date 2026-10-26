package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckOrigin(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := CheckOrigin([]string{"https://allowed.example"}, ok)

	cases := []struct {
		origin string
		want   int
	}{
		{"", http.StatusOK},                        // non-browser client
		{"http://game.example", http.StatusOK},     // same origin
		{"https://allowed.example", http.StatusOK}, // allow-listed
		{"https://evil.example", http.StatusForbidden},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, "http://game.example/api/ws", nil)
		if tc.origin != "" {
			req.Header.Set("Origin", tc.origin)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Errorf("origin %q: status %d, want %d", tc.origin, rec.Code, tc.want)
		}
	}
}

func TestRateLimiterBurstThenBlocks(t *testing.T) {
	l := NewRateLimiter(60, 3, false)
	for i := 0; i < 3; i++ {
		if ok, _ := l.Allow("1.2.3.4"); !ok {
			t.Fatalf("request %d blocked inside burst", i+1)
		}
	}
	if ok, wait := l.Allow("1.2.3.4"); ok || wait <= 0 {
		t.Fatalf("4th request allowed=%v wait=%v, want blocked", ok, wait)
	}
	if ok, _ := l.Allow("5.6.7.8"); !ok {
		t.Fatal("other client should have its own bucket")
	}
}

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:5555"
	req.Header.Set("X-Forwarded-For", "6.6.6.6, 203.0.113.9")

	if got := ClientIP(req, false); got != "10.0.0.1" {
		t.Errorf("untrusted proxy: got %q", got)
	}
	if got := ClientIP(req, true); got != "203.0.113.9" {
		t.Errorf("trusted proxy: got %q, want last hop", got)
	}
}
