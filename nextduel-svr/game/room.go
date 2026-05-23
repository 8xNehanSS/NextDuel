package game

import (
	"fmt"
	"nextduel-svr/models"
	"nextduel-svr/utils"
	"time"
)

type Room struct {
	ID      string
	Players map[string]*models.Player
	Scores  map[string]int
	Quest   models.Quest
	Status  string

	round             int
	maxRounds         int
	timer             uint
	intermissionTimer int
	fastDuelTimer     int
	onFastDuel        bool
	stopTimer         chan struct{}
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
		Status:            "created",
		round:             1,
		maxRounds:         5,
		timer:             61,
		intermissionTimer: 6,
		fastDuelTimer:     11,
		onFastDuel:        false,
		stopTimer:         make(chan struct{}),
	}
}

func (r *Room) TryStartGame() {
	if r.AreAllPlayersReady() && r.Status == "created" {
		r.Status = "active"

		go r.RoomTimer()
	}
}

func (r *Room) startGame() {
	r.Status = "active"
	for _, player := range r.Players {
		player.ClearAnswer()
		player.Conn.WriteJSON(map[string]interface{}{
			"type": utils.GameActionDuelStart,
			"payload": map[string]interface{}{
				"quest":   r.Quest.Question,
				"round":   r.round,
				"players": r.getPlayerInfo(),
			},
		})
	}
}

func (r *Room) prepareRound() {
	r.Quest = models.Quest{
		ID:       fmt.Sprintf("quest-%d", r.round),
		Question: fmt.Sprintf("What is the answer to question %d?", r.round),
		Answer:   "IDK",
	}
}

func (r *Room) RoomTimer() {

	for r.round <= r.maxRounds {
		r.prepareRound()
		r.startGame()

		ticker := time.NewTicker(1 * time.Second)

		remaining := r.timer

	roundLoop:
		for {
			select {
			case <-ticker.C:
				if r.onFastDuel && remaining > uint(r.fastDuelTimer) {
					remaining = uint(r.fastDuelTimer)
				}

				remaining--
				r.broadcastRoundTime(remaining)

				if remaining <= 0 {
					ticker.Stop()
					r.ProcessDuelResults()
					break roundLoop
				}

				if r.hasAllPlayersAnswered() {
					ticker.Stop()
					r.ProcessDuelResults()
					break roundLoop
				}

			case <-r.stopTimer:
				ticker.Stop()
				return
			}
		}

		ticker.Stop()
		r.round++
		r.resetRoundState()

		r.intermissionStart()

		intermission := r.intermissionTimer
		interTicker := time.NewTicker(1 * time.Second)

	intermissionLoop:
		for {
			select {
			case <-interTicker.C:

				intermission--
				r.broadcastIntermissionTime(intermission)

				if intermission <= 0 {
					interTicker.Stop()
					break intermissionLoop
				}

			case <-r.stopTimer:
				interTicker.Stop()
				return
			}
		}
	}

	r.endDuel()
}

func (r *Room) intermissionStart() {
	for _, player := range r.Players {
		player.Conn.WriteJSON(map[string]interface{}{
			"type":    utils.GameActionIntermission,
			"payload": nil,
		})
	}
}

func (r *Room) broadcastRoundTime(remaining uint) {
	for _, player := range r.Players {
		player.Conn.WriteJSON(map[string]interface{}{
			"type":    utils.GameActionTimerUpdate,
			"payload": remaining,
		})
	}
}

func (r *Room) broadcastIntermissionTime(remaining int) {
	for _, player := range r.Players {
		player.Conn.WriteJSON(map[string]interface{}{
			"type":    utils.GameActionIntermissionTimerUpdate,
			"payload": remaining,
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
		r.onFastDuel = true
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

func (r *Room) resetRoundState() {
	r.onFastDuel = false
	for _, player := range r.Players {
		player.ClearAnswer()
	}
}

func (r *Room) endDuel() error {
	// log to database
	r.Status = "completed"
	for _, player := range r.Players {
		player.Conn.WriteJSON(map[string]interface{}{
			"type":    utils.GameActionGameOver,
			"payload": r.getPlayerInfo(),
		})
	}
	return nil
}
