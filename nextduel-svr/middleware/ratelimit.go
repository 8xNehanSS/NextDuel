package middleware

import (
	"net"
	"net/http"
	"nex-server/server"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimiter is a per-client token bucket for nex-server handlers.
type RateLimiter struct {
	rate       float64 // tokens per second
	burst      float64
	trustProxy bool

	mu      sync.Mutex
	buckets map[string]*bucket
	lastGC  time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
}

// NewRateLimiter allows `perMinute` requests per minute with bursts of `burst`.
func NewRateLimiter(perMinute, burst int, trustProxy bool) *RateLimiter {
	return &RateLimiter{
		rate:       float64(perMinute) / 60,
		burst:      float64(burst),
		trustProxy: trustProxy,
		buckets:    make(map[string]*bucket),
		lastGC:     time.Now(),
	}
}

// Allow reports whether key may proceed, and if not, how long to wait.
func (l *RateLimiter) Allow(key string) (bool, time.Duration) {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	l.gcLocked(now)
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	b.tokens = min(l.burst, b.tokens+now.Sub(b.last).Seconds()*l.rate)
	b.last = now
	if b.tokens >= 1 {
		b.tokens--
		return true, 0
	}
	wait := time.Duration((1 - b.tokens) / l.rate * float64(time.Second))
	return false, wait
}

// Wrap guards a nex-server handler.
func (l *RateLimiter) Wrap(h server.Handler) server.Handler {
	return func(ctx *server.Context) error {
		ok, wait := l.Allow(ClientIP(ctx.Request(), l.trustProxy))
		if !ok {
			ctx.SetHeader("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
			return ctx.Error(http.StatusTooManyRequests, "Too many attempts. Please wait and try again.", nil)
		}
		return h(ctx)
	}
}

// Buckets idle long enough to be full again carry no state worth keeping.
func (l *RateLimiter) gcLocked(now time.Time) {
	if now.Sub(l.lastGC) < time.Minute {
		return
	}
	l.lastGC = now
	idle := time.Duration(l.burst/l.rate*float64(time.Second)) + time.Minute
	for k, b := range l.buckets {
		if now.Sub(b.last) > idle {
			delete(l.buckets, k)
		}
	}
}

// ClientIP returns the caller's address. With trustProxy it uses the last
// X-Forwarded-For hop, which is the one appended by our own proxy.
func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
