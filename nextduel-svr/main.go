package main

import (
	"fmt"
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

func main() {
	db, err := openDatabase()
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	g := gamepkg.NewGameManager(db)
	s := server.NewServer(db)
	s.SetService(g)
	// s.SetAuthFunc(utils.AuthenticateFunction)
	// s.SetWSAuthFunc(utils.Auth)
	s.Register(registerEndpoints)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}

func registerEndpoints(r *server.Registrar, config serverutils.Config) {
	api := r.WithPrefix("/api")

	// public routes
	api.Public(http.MethodPost, "/login", login.LoginHandler)
	api.Public(http.MethodPost, "/register", register.RegisterUser)

	// private routes
	api.Private(http.MethodGet, "/health", health.HealthHandler)
	api.Private(http.MethodGet, "/user", user.GetUser)
	api.Private(http.MethodPost, "/logout", login.HandleLogout)
	api.Private(http.MethodGet, "/history/duels", history.GetDuelHistory)

	// websocket routes
	api.PrivateWS("/ws", registerWSHandlers, gamepkg.HandleWSDisconnect)

	// file server
	api.ServeApplication(http.Dir(fmt.Sprint("./%s", config.HtmlRoot)))
}

func registerWSHandlers(ctx *server.Context, conn *websocket.Conn, msg server.WSMessage) error {
	player := gamepkg.NewPlayer(ctx.UserID(), conn)
	switch utils.GameAction(msg.Type) {

	case utils.GameActionPing:
		ctx.SendWSResponse(conn, utils.GameActionPong, "ok")

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
	return nil
}
