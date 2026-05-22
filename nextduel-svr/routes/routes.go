package routes

import (
	"log"
	"net/http"
	"nextduel-svr/models"
	game_events "nextduel-svr/routes/game-events"
	"nextduel-svr/routes/health"
	"nextduel-svr/routes/login"
	"nextduel-svr/routes/register"
	"nextduel-svr/server"
)

func RegisterRoutes(s *server.Server) {
	mux := s.Router()

	// HTML File server
	fileServer := http.FileServer(http.Dir("./web/dist"))
	mux.Handle("/", fileServer)

	// api routes
	apiMux := http.NewServeMux()
	registerPublicRoutes(s, apiMux)
	registerPrivateRoutes(s, apiMux)
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// websocket route
	mux.Handle("/ws", s.AuthMiddleware(WSHandler))
}

func registerPrivateRoutes(s *server.Server, mux *http.ServeMux) {
	mux.Handle("GET /health", s.AuthMiddleware(health.HealthHandler))
}

func registerPublicRoutes(s *server.Server, mux *http.ServeMux) {
	mux.Handle("POST /login", s.PublicMiddleware(login.LoginHandler))
	mux.Handle("POST /register", s.PublicMiddleware(register.RegisterUser))
}

func registerWSHandlers(ctx *server.Context, player *models.Player, msg WSMessage) {

	switch msg.Type {

	case "ping":
		ctx.SendWSResponse([]*models.Player{player}, "pong", "ok")

	case "join_queue":
		game_events.HandleJoinQueue(ctx, player, msg.Data)

	case "player_ready":
		game_events.HandleReadyPlayer(ctx, player, msg.Data)

	case "create_room":
		game_events.HandleCreateRoom(ctx, player, msg.Data)

	case "join_room":
		game_events.HandleJoinRoom(ctx, player, msg.Data)

	default:
		log.Println("Unknown message:", msg.Type)
	}
}
