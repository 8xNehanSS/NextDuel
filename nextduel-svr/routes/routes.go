package routes

import (
	"log"
	"net/http"
	"nextduel-svr/models"
	game_events "nextduel-svr/routes/game-events"
	"nextduel-svr/routes/health"
	"nextduel-svr/routes/login"
	"nextduel-svr/routes/register"
	"nextduel-svr/routes/user"
	"nextduel-svr/server"
	"nextduel-svr/utils"
)

func RegisterRoutes(s *server.Server) {
	mux := s.Router()

	// HTML File server
	fileServer := http.FileServer(http.Dir("./htmls"))
	mux.Handle("/", fileServer)

	// api routes
	apiMux := http.NewServeMux()
	registerPublicRoutes(s, apiMux)
	registerPrivateRoutes(s, apiMux)
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// websocket route
	mux.Handle("/ws", s.WsMiddleware(WSHandler))
}

func registerPrivateRoutes(s *server.Server, mux *http.ServeMux) {
	mux.Handle("GET /health", s.AuthMiddleware(health.HealthHandler))
	mux.Handle("GET /user", s.AuthMiddleware(user.GetUser))
	mux.Handle("POST /logout", s.AuthMiddleware(login.HandleLogout))
}

func registerPublicRoutes(s *server.Server, mux *http.ServeMux) {
	mux.Handle("POST /login", s.PublicMiddleware(login.LoginHandler))
	mux.Handle("POST /register", s.PublicMiddleware(register.RegisterUser))
}

func registerWSHandlers(ctx *server.Context, player *models.Player, msg WSMessage) {

	switch utils.GameAction(msg.Type) {

	case utils.GameActionPing:
		ctx.SendWSResponse([]*models.Player{player}, utils.GameActionPong, "ok")

	case utils.GameActionJoinQueue:
		game_events.HandleJoinQueue(ctx, player, msg.Data)

	case utils.GameActionPlayerReady:
		game_events.HandleReadyPlayer(ctx, player, msg.Data)

	case utils.GameActionCreateRoom:
		game_events.HandleCreateRoom(ctx, player, msg.Data)

	case utils.GameActionJoinRoom:
		game_events.HandleJoinRoom(ctx, player, msg.Data)

	case utils.GameActionDuelAnswer:
		game_events.HandleDuelAnswer(ctx, player, msg.Data)

	default:
		log.Println("Unknown message:", msg.Type)
	}
}
