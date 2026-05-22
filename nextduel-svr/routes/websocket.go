package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"nextduel-svr/models"
	"nextduel-svr/server"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func WSHandler(ctx *server.Context) error {

	conn, err := upgrader.Upgrade(ctx.ResponseWriter(), ctx.Request(), nil)
	if err != nil {
		log.Println("WS error:", err)
		return err
	}

	player := &models.Player{
		ID:   ctx.GetCurrentUserID(),
		Conn: conn,
	}

	log.Println("Player connected:", player.ID)

	go readLoop(ctx, player)
	return nil
}

func readLoop(ctx *server.Context, player *models.Player) {

	defer func() {
		player.Conn.Close()
		log.Println("Player disconnected:", player.ID)
	}()

	for {

		var msg WSMessage

		err := player.Conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		registerWSHandlers(ctx, player, msg)
	}
}
