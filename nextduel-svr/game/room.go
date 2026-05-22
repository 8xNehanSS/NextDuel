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

	counter int
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
		Status:  "created",
		counter: 1,
	}
}

func (r *Room) StartGame() {
	r.Status = "active"
	for _, player := range r.Players {
		player.ClearAnswer()
		player.Conn.WriteJSON(map[string]interface{}{
			"type":    utils.GameActionDuelStart,
			"quest":   r.Quest.Question,
			"round":   r.counter,
			"players": r.getPlayerInfo(),
		})
	}
}

func (r *Room) getPlayerInfo() []map[string]interface{} {
	info := []map[string]interface{}{}
	for _, player := range r.Players {
		info = append(info, map[string]interface{}{
			"id":    player.ID,
			"score": player.GetScore(),
		})
	}
	return info
}

func (r *Room) AreAllPlayersReady() bool {

	if len(r.Players) == 0 {
		return false
	}

	for _, p := range r.Players {
		if !p.IsReady() {
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

func (r *Room) GetOpponent(playerID string) *models.Player {
	for id, player := range r.Players {
		if id != playerID {
			return player
		}
	}
	return nil
}

func (r *Room) GetPlayers() []*models.Player {
	players := []*models.Player{}
	for _, player := range r.Players {
		players = append(players, player)
	}
	return players
}

func (r *Room) AnswerDuel(playerID string, data interface{}) (bool, error) {
	if player, exists := r.Players[playerID]; exists {
		if player.GetAnswer() != nil {
			return false, fmt.Errorf("player has already answered")
		}
		player.SetAnswer(data)
	}
	hasAllAnswered := r.hasAllPlayersAnswered()
	if !hasAllAnswered {
		return false, nil
	}
	return true, nil
}

func (r *Room) hasAllPlayersAnswered() bool {
	for _, player := range r.Players {
		if player.GetAnswer() == nil {
			return false
		}
	}
	return true
}

func (r *Room) ProcessDuelResults() {
	// Placeholder for processing duel results and updating scores
	// log in database
}

func (r *Room) NextDuel() {
	if r.Status != "active" {
		return
	}

	if !r.AreAllPlayersReady() {
		return
	}

	if r.counter >= 5 {
		r.CompleteDuel()
		return
	}

	r.Quest = models.Quest{
		ID:       fmt.Sprintf("quest-%d", r.counter),
		Question: fmt.Sprintf("What is %d + %d?", r.counter, r.counter*2),
		Answer:   fmt.Sprintf("%d", r.counter*3),
	}

	r.counter++
	r.StartGame()
}

func (r *Room) CompleteDuel() error {
	// log to database
	r.Status = "completed"
	for _, player := range r.Players {
		player.Conn.WriteJSON(map[string]interface{}{
			"type":    utils.GameActionGameOver,
			"players": r.getPlayerInfo(),
		})
	}
	return nil
}
