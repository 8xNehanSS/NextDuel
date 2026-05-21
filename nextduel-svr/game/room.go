package game

import "nextduel-svr/models"

type Room struct {
	ID      string
	Players map[string]*models.Player
	Scores  map[string]int
	Quest   models.Quest
	Status  string
}

func (r *Room) startGame() {
	r.Status = "active"
	// Send initial game state to players
	for _, player := range r.Players {
		player.Conn.WriteJSON(map[string]interface{}{
			"type":    "game_start",
			"quest":   r.Quest.Question,
			"players": r.getPlayerInfo(),
		})
	}
}

func (r *Room) getPlayerInfo() []map[string]interface{} {
	info := []map[string]interface{}{}
	for _, player := range r.Players {
		info = append(info, map[string]interface{}{
			"id":    player.ID,
			"score": player.Score,
		})
	}
	return info
}
