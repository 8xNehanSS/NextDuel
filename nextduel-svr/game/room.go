package game

import (
	"fmt"
	"nextduel-svr/models"
	"nextduel-svr/utils"
)

type Room struct {
	ID      string
	Players map[string]*models.Player
	Scores  map[string]int
	Quest   models.Quest
	Status  string
}

func NewRoom(p1, p2 *models.Player) *Room {
	roomID := utils.GenerateRoomID()
	return &Room{
		ID: roomID,
		Players: map[string]*models.Player{
			p1.ID: p1,
			p2.ID: p2,
		},
		Scores: map[string]int{
			p1.ID: 0,
			p2.ID: 0,
		},
		Status: "created",
	}
}

func (r *Room) StartGame() {
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

func (r *Room) AreAllPlayersReady() bool {

	if len(r.Players) == 0 {
		return false
	}

	for _, p := range r.Players {
		if !p.Ready {
			return false
		}
	}

	return true
}

func (r *Room) SetPlayerReady(playerID string) {
	if player, exists := r.Players[playerID]; exists {
		player.SetReady(true)
	}
}

func (r *Room) AddPlayer(p *models.Player) error {
	if len(r.Players) >= 2 {
		return fmt.Errorf("room is full")
	}
	r.Players[p.ID] = p
	r.Scores[p.ID] = 0
	return nil
}

func (r *Room) GetPlayers() []*models.Player {
	players := []*models.Player{}
	for _, player := range r.Players {
		players = append(players, player)
	}
	return players
}
