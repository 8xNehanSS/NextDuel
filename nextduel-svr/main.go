package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"nex-server/server"
	"nextduel-svr/config"
	"nextduel-svr/database"
	"nextduel-svr/gamepkg"
	"nextduel-svr/middleware"
	"nextduel-svr/questions"
	"nextduel-svr/repository"
	game_events "nextduel-svr/routes/game-events"
	"nextduel-svr/routes/health"
	"nextduel-svr/routes/history"
	"nextduel-svr/routes/leaderboard"
	"nextduel-svr/routes/login"
	"nextduel-svr/routes/register"
	"nextduel-svr/routes/user"
	"os"
	"os/signal"
	"syscall"
	"time"

	serverutils "nex-server/utils"

	"gorm.io/gorm"
)

// version is overridden at build time: -ldflags "-X main.version=1.2.3".
var version = "dev"

type Services struct {
	Server *server.Server
	Game   *gamepkg.GameManager
	Config config.Config
}

func main() {
	cfg := config.Load()

	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(healthcheck(cfg.Port))
	}

	setupLogging(cfg)
	if err := run(cfg); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	handler, game := newApp(cfg, db)

	// nex-server's Start() offers no timeouts or graceful shutdown, so we run
	// its router on our own http.Server.
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server listening", "addr", srv.Addr, "version", version)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		slog.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Hijacked websocket connections are not tracked by http.Server, so the
	// game manager notifies and closes them itself.
	game.Shutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Warn("http shutdown", "error", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	return nil
}

// newApp wires the game manager and every route. nex-server's CORS
// middleware is deliberately not used: the SPA is served same-origin.
func newApp(cfg config.Config, db *gorm.DB) (http.Handler, *gamepkg.GameManager) {
	game := gamepkg.NewGameManager(gameSettings(cfg), newModes(cfg, db), repository.NewMatchRecorder(db))
	s := server.NewServer(db)
	s.Register(registerEndpoints, &Services{Server: s, Game: game, Config: cfg})

	handler := middleware.Chain(s.Router(),
		middleware.Recover,
		middleware.AccessLog,
		middleware.SecurityHeaders,
	)
	return handler, game
}

func registerEndpoints(r *server.Registrar, config serverutils.Config, svc any) {
	services := svc.(*Services)
	api := r.WithPrefix("/api")
	// per client IP: 10/min sustained, bursts of 5
	loginLimiter := middleware.NewRateLimiter(10, 5, services.Config.TrustProxy)
	registerLimiter := middleware.NewRateLimiter(10, 5, services.Config.TrustProxy)

	// public routes
	api.Public(http.MethodGet, "/health", health.New(services.Game, version).HealthHandler)
	api.Public(http.MethodPost, "/login", loginLimiter.Wrap(login.LoginHandler))
	api.Public(http.MethodPost, "/register", registerLimiter.Wrap(register.RegisterUser))

	// private routes
	api.Private(http.MethodGet, "/stats", health.New(services.Game, version).StatsHandler)
	api.Private(http.MethodGet, "/user", user.GetUser)
	api.Private(http.MethodPost, "/logout", login.HandleLogout)
	api.Private(http.MethodGet, "/history/duels", history.GetDuelHistory)
	api.Private(http.MethodGet, "/leaderboard", leaderboard.GetLeaderboard)

	// websocket: registered by hand so the origin check runs before upgrade
	events := game_events.New(services.Game)
	r.Handle("GET /api/ws", middleware.CheckOrigin(services.Config.AllowedOrigins,
		services.Server.WsAuthMiddleware(events.HandleMessage, events.HandleDisconnect, services)))

	// unknown API paths must 404 rather than fall through to the SPA
	r.Handle("/api/", http.NotFoundHandler())

	// single-page app
	r.Handle("/", middleware.SPA(config.HtmlRoot))
}

func gameSettings(cfg config.Config) gamepkg.Settings {
	s := gamepkg.DefaultSettings()
	s.Rounds = cfg.Game.Rounds
	s.RoundDuration = cfg.Game.RoundDuration
	s.FastDuelDuration = cfg.Game.FastDuelDuration
	s.IntermissionDuration = cfg.Game.IntermissionDuration
	return s
}

// newModes registers every playable game. Adding a game = implement
// gamepkg.Mode, add its ID to utils.GameMode, register it here and add it
// to nextduel-frontend/src/games.ts.
func newModes(cfg config.Config, db *gorm.DB) gamepkg.Modes {
	return gamepkg.NewModes(
		gamepkg.NewTriviaMode(questionProvider(cfg, db)),
		gamepkg.NewMathMode(),
		gamepkg.NewScrambleMode(),
		gamepkg.NewRPSMode(),
	)
}

func questionProvider(cfg config.Config, db *gorm.DB) questions.Provider {
	bank := questions.NewBank(db)
	if !cfg.OpenTDBEnabled {
		return bank
	}
	return &questions.Fallback{
		Primary:   questions.NewOpenTDB(),
		Secondary: bank,
		Cache:     bank.CacheFunc("opentdb"),
	}
}

func setupLogging(cfg config.Config) {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	var h slog.Handler = slog.NewTextHandler(os.Stdout, opts)
	if cfg.LogFormat == "json" {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	// Also routes nex-server's log.Printf output through slog.
	slog.SetDefault(slog.New(h))
}

func healthcheck(port string) int {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/api/health")
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck:", err)
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "healthcheck: status", resp.StatusCode)
		return 1
	}
	return 0
}
