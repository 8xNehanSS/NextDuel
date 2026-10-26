// Package config loads application settings from the environment.
//
// Framework-level settings (PORT, COOKIE_KEY, HTML_ROOT, SERVER_ROOT) are read
// by nex-server itself; this package covers everything specific to NextDuel.
// See .env.example at the repository root for the full list.
package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type GameConfig struct {
	Rounds               int
	RoundDuration        time.Duration
	FastDuelDuration     time.Duration
	IntermissionDuration time.Duration
}

type Config struct {
	Port string

	DBPath string
	// AdminPassword seeds an "admin" account on first start. Empty = no seed.
	AdminPassword string

	// TrustProxy makes the rate limiter key on X-Forwarded-For. Enable only
	// when the server is reachable exclusively through a reverse proxy.
	TrustProxy bool
	// AllowedOrigins is an extra allow-list for websocket Origin checks.
	// Same-origin requests are always allowed.
	AllowedOrigins []string

	OpenTDBEnabled bool

	LogFormat string // "text" or "json"
	LogLevel  slog.Level

	Game GameConfig
}

// Load reads .env (if present) and the process environment.
func Load() Config {
	_ = godotenv.Load()

	return Config{
		Port:           str("PORT", "8080"),
		DBPath:         str("DB_PATH", "nextduel.db"),
		AdminPassword:  os.Getenv("ADMIN_PASSWORD"),
		TrustProxy:     boolean("TRUST_PROXY", false),
		AllowedOrigins: list("ALLOWED_ORIGINS"),
		OpenTDBEnabled: boolean("OPENTDB_ENABLED", true),
		LogFormat:      str("LOG_FORMAT", "text"),
		LogLevel:       level("LOG_LEVEL", slog.LevelInfo),
		Game: GameConfig{
			Rounds:               integer("GAME_ROUNDS", 5),
			RoundDuration:        seconds("GAME_ROUND_SECONDS", 60),
			FastDuelDuration:     seconds("GAME_FAST_DUEL_SECONDS", 10),
			IntermissionDuration: seconds("GAME_INTERMISSION_SECONDS", 5),
		},
	}
}

func str(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func boolean(key string, fallback bool) bool {
	v, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return v
}

func integer(key string, fallback int) int {
	v, err := strconv.Atoi(os.Getenv(key))
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func seconds(key string, fallback int) time.Duration {
	return time.Duration(integer(key, fallback)) * time.Second
}

func list(key string) []string {
	var out []string
	for _, part := range strings.Split(os.Getenv(key), ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func level(key string, fallback slog.Level) slog.Level {
	var l slog.Level
	if err := l.UnmarshalText([]byte(os.Getenv(key))); err != nil {
		return fallback
	}
	return l
}
