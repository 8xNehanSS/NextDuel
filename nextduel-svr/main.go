package main

import (
	"log"
	"net/http"
	"nex-server/server"
	"nextduel-svr/gamepkg"
	game_events "nextduel-svr/routes/game-events"
	"nextduel-svr/routes/health"
	"nextduel-svr/routes/history"
	"nextduel-svr/routes/login"
	"nextduel-svr/routes/register"
	"nextduel-svr/routes/user"
	"nextduel-svr/utils"

	serverutils "nex-server/utils"

	"github.com/gorilla/websocket"
)

type Services struct {
	Game *gamepkg.GameManager
}

func main() {
	db, err := openDatabase()
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	g := gamepkg.NewGameManager(db)
	s := server.NewServer(db)

	services := &Services{
		Game: g,
	}
	// s.SetAuthFunc(utils.AuthenticateFunction)
	// s.SetWSAuthFunc(utils.Auth)

	s.Register(registerEndpoints, services)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}

func registerEndpoints(r *server.Registrar, config serverutils.Config, services any) {
	api := r.WithPrefix("/api")

	// public routes
	api.Public(http.MethodPost, "/login", login.LoginHandler)
	api.Public(http.MethodPost, "/register", register.RegisterUser)

	// private routes
	// health handlers
	healthSvc := health.New(services.(*Services).Game)
	api.Private(http.MethodGet, "/health", healthSvc.HealthHandler)

	api.Private(http.MethodGet, "/user", user.GetUser)
	api.Private(http.MethodPost, "/logout", login.HandleLogout)
	api.Private(http.MethodGet, "/history/duels", history.GetDuelHistory)

	// websocket routes
	api.PrivateWS("/ws", registerWSHandlers, services.(*Services).Game.HandleWSDisconnect, services)

	// file server
	r.Handle("/", r.ServeApplication(http.Dir(config.HtmlRoot)))
}

func registerWSHandlers(ctx *server.Context, conn *websocket.Conn, msg server.WSMessage, services any) error {
	player := gamepkg.NewPlayer(ctx.UserID(), conn)
	gameEventSvc := game_events.New(services.(*Services).Game)
	switch utils.GameAction(msg.Type) {

	case utils.GameActionPing:
		ctx.SendWSResponse(conn, utils.GameActionPong, "ok")

	case utils.GameActionJoinQueue:
		gameEventSvc.HandleJoinQueue(ctx, player, msg.Data)

	case utils.GameActionPlayerReady:
		gameEventSvc.HandleReadyPlayer(ctx, player, msg.Data)

	case utils.GameActionCreateRoom:
		gameEventSvc.HandleCreateRoom(ctx, player, msg.Data)

	case utils.GameActionJoinRoom:
		gameEventSvc.HandleJoinRoom(ctx, player, msg.Data)

	case utils.GameActionDuelAnswer:
		gameEventSvc.HandleDuelAnswer(ctx, player, msg.Data)

	default:
		log.Println("Unknown message:", msg.Type)
	}
	return nil
}
