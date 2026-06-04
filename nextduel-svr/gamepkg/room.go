package gamepkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"nextduel-svr/models"
	"nextduel-svr/utils"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type Room struct {
	ID      string
	Players map[string]*Player
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
	destroyOnce       sync.Once
	onRoomDestroyed   func(roomID string)

	db              *gorm.DB
	matchID         uint
	roundID         uint
	playerRecordIDs map[string]uint
	roundWins       map[string]int
	playerOrder     []string
}

func NewRoom(p1, p2 *Player, db *gorm.DB) *Room {
	roomID := utils.GenerateRoomID()
	players := map[string]*Player{
		p1.ID: p1,
	}
	scores := map[string]int{
		p1.ID: 0,
	}
	playerOrder := []string{p1.ID}
	if p2 != nil {
		players[p2.ID] = p2
		scores[p2.ID] = 0
		playerOrder = append(playerOrder, p2.ID)
	}

	room := &Room{
		ID:                roomID,
		Players:           players,
		Scores:            scores,
		Status:            string(utils.RoomStatusCreated),
		round:             1,
		maxRounds:         5,
		timer:             60,
		intermissionTimer: 5,
		fastDuelTimer:     10,
		onFastDuel:        false,
		stopTimer:         make(chan struct{}),
		db:                db,
		playerRecordIDs:   make(map[string]uint),
		roundWins:         make(map[string]int),
		playerOrder:       playerOrder,
	}

	for _, playerID := range playerOrder {
		room.roundWins[playerID] = 0
	}

	room.initMatch()
	return room
}

func (r *Room) initMatch() {
	if r.db == nil || len(r.playerOrder) == 0 {
		return
	}

	creatorID, err := r.resolvePlayerRecord(r.playerOrder[0])
	if err != nil {
		log.Printf("match init: failed to resolve creator: %v", err)
		return
	}

	match := models.Match{
		Mode:      string(utils.MatchModeClassic),
		Status:    string(utils.MatchStatusWaiting),
		BestOf:    r.maxRounds,
		CreatedBy: creatorID,
	}
	if err := r.db.Create(&match).Error; err != nil {
		log.Printf("match init: failed to create match: %v", err)
		return
	}

	r.matchID = match.ID
	r.logMatchEvent(utils.MatchEventCreated, nil, map[string]any{"room_id": r.ID})

	for slot, playerID := range r.playerOrder {
		recordID, err := r.resolvePlayerRecord(playerID)
		if err != nil {
			log.Printf("match init: failed to resolve player: %v", err)
			continue
		}
		matchPlayer := models.MatchPlayer{
			MatchID:  r.matchID,
			PlayerID: recordID,
			Slot:     slot + 1,
			JoinedAt: time.Now(),
		}
		if err := r.db.Create(&matchPlayer).Error; err != nil {
			log.Printf("match init: failed to create match player: %v", err)
			continue
		}
		r.logMatchEvent(utils.MatchEventPlayerJoin, &recordID, map[string]any{"slot": slot + 1})
	}
}

func (r *Room) TryStartGame() {
	if r.AreAllPlayersReady() && r.Status == string(utils.RoomStatusCreated) {
		r.Status = string(utils.RoomStatusActive)
		r.markMatchStarted()

		go r.RoomTimer()
	}
}

func (r *Room) startGame() {
	r.Status = string(utils.RoomStatusActive)
	r.createRoundRecord()
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

				if remaining == 0 {
					ticker.Stop()
					r.ProcessDuelResults()
					break roundLoop
				}

				remaining--
				r.broadcastRoundTime(remaining)

				if remaining == 0 {
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
	if r.Status != string(utils.RoomStatusCreated) {
		return
	}
	if player, exists := r.Players[playerID]; exists {
		player.SetReady(true)
		recordID, err := r.resolvePlayerRecord(playerID)
		if err == nil {
			r.logMatchEvent(utils.MatchEventPlayerReady, &recordID, map[string]any{"round": r.round})
		}
	}
}

func (r *Room) AddPlayer(p *Player) error {
	if len(r.Players) >= 2 {
		return fmt.Errorf("room is full")
	}
	r.Players[p.ID] = p
	r.Scores[p.ID] = 0
	r.roundWins[p.ID] = 0
	r.playerOrder = append(r.playerOrder, p.ID)
	r.logPlayerJoin(p.ID, len(r.playerOrder))
	if len(r.Players) == 2 {
		r.markMatchReady()
	}
	return nil
}

func (r *Room) GetOpponent(playerID string) *Player {
	for id, player := range r.Players {
		if id != playerID {
			return player
		}
	}
	return nil
}

func (r *Room) GetPlayers() []*Player {
	players := []*Player{}
	for _, player := range r.Players {
		players = append(players, player)
	}
	return players
}

func (r *Room) GetPlayerCons() []*websocket.Conn {
	conns := []*websocket.Conn{}
	for _, player := range r.Players {
		conns = append(conns, player.Conn)
	}
	return conns
}

func (r *Room) AnswerDuel(playerID string, data interface{}) (bool, error) {
	if r.Status != string(utils.RoomStatusActive) {
		return false, fmt.Errorf("match is not active")
	}
	if player, exists := r.Players[playerID]; exists {
		if player.GetAnswer() != nil {
			return false, fmt.Errorf("player has already answered")
		}
		player.SetAnswer(data)
		r.logRoundAnswer(playerID, data)
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
	if len(r.Players) == 0 {
		return
	}

	results := make(map[string]int)
	correctAnswer := strings.TrimSpace(r.Quest.Answer)
	winnerPlayerID := ""
	maxDelta := -1

	for _, player := range r.Players {
		answer := r.normalizeAnswer(player.GetAnswer())
		delta := 0
		if answer != "" && strings.EqualFold(answer, correctAnswer) {
			delta = 1
		}
		results[player.ID] = delta
		if delta > maxDelta {
			maxDelta = delta
			winnerPlayerID = player.ID
		} else if delta == maxDelta {
			winnerPlayerID = ""
		}
		player.IncrementScore(delta)
		r.Scores[player.ID] += delta
	}

	var winnerRecordID *uint
	if winnerPlayerID != "" {
		id, err := r.resolvePlayerRecord(winnerPlayerID)
		if err == nil {
			winnerRecordID = &id
		}
	}

	if r.db != nil && r.matchID != 0 {
		update := map[string]any{
			"ended_at": time.Now(),
		}
		if winnerRecordID != nil {
			update["winner_id"] = winnerRecordID
		}
		if r.roundID != 0 {
			if err := r.db.Model(&models.MatchRound{}).Where("id = ?", r.roundID).Updates(update).Error; err != nil {
				log.Printf("round result: failed to update match round: %v", err)
			}
		}
	}

	for _, player := range r.Players {
		recordID, err := r.resolvePlayerRecord(player.ID)
		if err != nil {
			log.Printf("round result: failed to resolve player: %v", err)
			continue
		}

		delta := results[player.ID]
		isWinner := winnerPlayerID != "" && player.ID == winnerPlayerID
		if isWinner {
			r.roundWins[player.ID]++
		}

		if r.db != nil && r.roundID != 0 {
			var entry models.RoundPlayer
			err := r.db.Where("round_id = ? AND player_id = ?", r.roundID, recordID).First(&entry).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					entry = models.RoundPlayer{
						RoundID:    r.roundID,
						PlayerID:   recordID,
						ScoreDelta: delta,
						IsWinner:   isWinner,
						Answer:     r.normalizeAnswer(player.GetAnswer()),
						AnsweredAt: time.Now(),
					}
					if err := r.db.Create(&entry).Error; err != nil {
						log.Printf("round result: failed to create round player: %v", err)
					}
				} else {
					log.Printf("round result: failed to lookup round player: %v", err)
				}
			} else {
				entry.ScoreDelta = delta
				entry.IsWinner = isWinner
				if err := r.db.Save(&entry).Error; err != nil {
					log.Printf("round result: failed to update round player: %v", err)
				}
			}
		}

		if r.db != nil && r.matchID != 0 {
			updates := map[string]any{
				"score": gorm.Expr("score + ?", delta),
			}
			if isWinner {
				updates["rounds_won"] = gorm.Expr("rounds_won + ?", 1)
			}
			if err := r.db.Model(&models.MatchPlayer{}).Where("match_id = ? AND player_id = ?", r.matchID, recordID).
				Updates(updates).Error; err != nil {
				log.Printf("round result: failed to update match player: %v", err)
			}
		}

		r.logScoreEvent(recordID, delta, r.Scores[player.ID], utils.ScoreEventRoundScore, utils.ScoreReasonRoundEnd)
	}

	r.logMatchEvent(utils.MatchEventRoundResult, winnerRecordID, map[string]any{"round": r.round, "scores": r.Scores})
}

func (r *Room) resetRoundState() {
	r.onFastDuel = false
	r.roundID = 0
	for _, player := range r.Players {
		player.ClearAnswer()
	}
}

func (r *Room) endDuel() error {
	matchWinner := r.resolveMatchWinner()
	if r.db != nil && r.matchID != 0 {
		updates := map[string]any{
			"status":   string(utils.MatchStatusCompleted),
			"ended_at": time.Now(),
		}
		if matchWinner != nil {
			updates["winner_id"] = matchWinner
		}
		if err := r.db.Model(&models.Match{}).Where("id = ?", r.matchID).Updates(updates).Error; err != nil {
			log.Printf("match end: failed to update match: %v", err)
		}

		if err := r.db.Model(&models.MatchPlayer{}).Where("match_id = ?", r.matchID).
			Updates(map[string]any{"left_at": time.Now()}).Error; err != nil {
			log.Printf("match end: failed to update match players: %v", err)
		}

		r.updatePlayerStats(matchWinner)
		r.logMatchEvent(utils.MatchEventDuelDone, matchWinner, map[string]any{"room_id": r.ID})
	}

	r.Status = string(utils.RoomStatusCompleted)
	for _, player := range r.Players {
		player.Conn.WriteJSON(map[string]interface{}{
			"type":    utils.GameActionGameOver,
			"payload": r.getPlayerInfo(),
		})
	}

	r.destroyRoom()
	return nil
}

func (r *Room) broadcast(gameAction utils.GameAction, data any) {
	for _, player := range r.Players {
		player.Conn.WriteJSON(map[string]interface{}{
			"type":    gameAction,
			"payload": data,
		})
	}
}

func (r *Room) HandleDisconnect(playerID string) {
	if r.db == nil || r.matchID == 0 {
		return
	}

	recordID, err := r.resolvePlayerRecord(playerID)
	if err != nil {
		log.Printf("disconnect: failed to resolve player: %v", err)
		return
	}

	if err := r.db.Model(&models.MatchPlayer{}).Where("match_id = ? AND player_id = ?", r.matchID, recordID).
		Updates(map[string]any{"left_at": time.Now()}).Error; err != nil {
		log.Printf("disconnect: failed to update match player: %v", err)
	}

	r.logMatchEvent(utils.MatchEventPlayerLeft, &recordID, map[string]any{"room_id": r.ID})

	opponent := r.GetOpponent(playerID)
	if opponent == nil {
		return
	}

	oppID, err := r.resolvePlayerRecord(opponent.ID)
	if err != nil {
		log.Printf("disconnect: failed to resolve opponent: %v", err)
		return
	}

	if err := r.db.Model(&models.Match{}).Where("id = ?", r.matchID).
		Updates(map[string]any{"status": string(utils.MatchStatusAbandoned), "winner_id": &oppID, "ended_at": time.Now()}).Error; err != nil {
		log.Printf("disconnect: failed to update match: %v", err)
	}

	r.updatePlayerStats(&oppID)
	r.logMatchEvent(utils.MatchEventDuelDone, &oppID, map[string]any{"reason": utils.ScoreReasonDisconnect})
}

func (r *Room) markMatchReady() {
	if r.db == nil || r.matchID == 0 {
		return
	}
	if err := r.db.Model(&models.Match{}).Where("id = ?", r.matchID).
		Updates(map[string]any{"status": string(utils.MatchStatusReady)}).Error; err != nil {
		log.Printf("match ready: failed to update match: %v", err)
		return
	}
	r.logMatchEvent(utils.MatchEventReady, nil, map[string]any{"room_id": r.ID})
}

func (r *Room) markMatchStarted() {
	if r.db == nil || r.matchID == 0 {
		return
	}
	if err := r.db.Model(&models.Match{}).Where("id = ?", r.matchID).
		Updates(map[string]any{"status": string(utils.MatchStatusActive), "started_at": time.Now()}).Error; err != nil {
		log.Printf("match start: failed to update match: %v", err)
		return
	}
	r.logMatchEvent(utils.MatchEventStarted, nil, map[string]any{"round": r.round})
}

func (r *Room) createRoundRecord() {
	if r.db == nil || r.matchID == 0 {
		return
	}
	round := models.MatchRound{
		MatchID:   r.matchID,
		RoundNo:   r.round,
		QuestID:   r.Quest.ID,
		Question:  r.Quest.Question,
		StartedAt: time.Now(),
	}
	if err := r.db.Create(&round).Error; err != nil {
		log.Printf("round start: failed to create match round: %v", err)
		return
	}
	r.roundID = round.ID
	r.logMatchEvent(utils.MatchEventRoundStart, nil, map[string]any{"round": r.round})
}

func (r *Room) logPlayerJoin(playerID string, slot int) {
	if r.db == nil || r.matchID == 0 {
		return
	}

	recordID, err := r.resolvePlayerRecord(playerID)
	if err != nil {
		log.Printf("player join: failed to resolve player: %v", err)
		return
	}

	var existing models.MatchPlayer
	lookup := r.db.Where("match_id = ? AND player_id = ?", r.matchID, recordID).First(&existing)
	if lookup.Error == nil {
		return
	}
	if lookup.Error != nil && !errors.Is(lookup.Error, gorm.ErrRecordNotFound) {
		log.Printf("player join: failed to lookup match player: %v", lookup.Error)
		return
	}

	matchPlayer := models.MatchPlayer{
		MatchID:  r.matchID,
		PlayerID: recordID,
		Slot:     slot,
		JoinedAt: time.Now(),
	}
	if err := r.db.Create(&matchPlayer).Error; err != nil {
		log.Printf("player join: failed to create match player: %v", err)
		return
	}

	r.logMatchEvent(utils.MatchEventPlayerJoin, &recordID, map[string]any{"slot": slot})
}

func (r *Room) logRoundAnswer(playerID string, answer any) {
	if r.db == nil || r.roundID == 0 {
		return
	}

	recordID, err := r.resolvePlayerRecord(playerID)
	if err != nil {
		log.Printf("answer log: failed to resolve player: %v", err)
		return
	}

	answerText := r.normalizeAnswer(answer)
	var entry models.RoundPlayer
	err = r.db.Where("round_id = ? AND player_id = ?", r.roundID, recordID).First(&entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			entry = models.RoundPlayer{
				RoundID:    r.roundID,
				PlayerID:   recordID,
				Answer:     answerText,
				AnsweredAt: time.Now(),
			}
			if err := r.db.Create(&entry).Error; err != nil {
				log.Printf("answer log: failed to create round player: %v", err)
			}
			r.logMatchEvent(utils.MatchEventDuelAnswer, &recordID, map[string]any{"round": r.round})
		} else {
			log.Printf("answer log: failed to lookup round player: %v", err)
		}
		return
	}

	entry.Answer = answerText
	entry.AnsweredAt = time.Now()
	if err := r.db.Save(&entry).Error; err != nil {
		log.Printf("answer log: failed to update round player: %v", err)
	}

	r.logMatchEvent(utils.MatchEventDuelAnswer, &recordID, map[string]any{"round": r.round})
}

func (r *Room) resolvePlayerRecord(username string) (uint, error) {
	if r.db == nil {
		return 0, fmt.Errorf("missing database")
	}
	if id, ok := r.playerRecordIDs[username]; ok {
		return id, nil
	}

	var player models.Players
	if err := r.db.Where("username = ?", username).First(&player).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
		player = models.Players{Username: username}
		if err := r.db.Create(&player).Error; err != nil {
			return 0, err
		}
	}

	r.playerRecordIDs[username] = player.ID
	return player.ID, nil
}

func (r *Room) normalizeAnswer(answer any) string {
	if answer == nil {
		return ""
	}
	switch value := answer.(type) {
	case string:
		return strings.TrimSpace(value)
	case []byte:
		var decoded string
		if err := json.Unmarshal(value, &decoded); err == nil {
			return strings.TrimSpace(decoded)
		}
		return strings.TrimSpace(string(value))
	default:
		payload, err := json.Marshal(value)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(payload))
	}
}

func (r *Room) logScoreEvent(playerID uint, delta int, total int, eventType utils.ScoreEventType, reason utils.ScoreEventReason) {
	if r.db == nil || r.matchID == 0 {
		return
	}

	entry := models.ScoreEvent{
		MatchID:   r.matchID,
		RoundID:   r.roundIDPtr(),
		PlayerID:  playerID,
		EventType: string(eventType),
		Reason:    string(reason),
		Delta:     delta,
		Total:     total,
		CreatedAt: time.Now(),
	}
	if err := r.db.Create(&entry).Error; err != nil {
		log.Printf("score event: failed to create entry: %v", err)
	}
}

func (r *Room) logMatchEvent(eventType utils.MatchEventType, playerID *uint, payload any) {
	if r.db == nil || r.matchID == 0 {
		return
	}

	entry := models.MatchEventLog{
		MatchID:    r.matchID,
		RoundID:    r.roundIDPtr(),
		PlayerID:   playerID,
		EventType:  string(eventType),
		Payload:    r.payloadString(payload),
		OccurredAt: time.Now(),
	}
	if err := r.db.Create(&entry).Error; err != nil {
		log.Printf("match event: failed to create entry: %v", err)
	}
}

func (r *Room) payloadString(payload any) string {
	if payload == nil {
		return ""
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(data)
}

func (r *Room) roundIDPtr() *uint {
	if r.roundID == 0 {
		return nil
	}
	value := r.roundID
	return &value
}

func (r *Room) destroyRoom() {
	r.destroyOnce.Do(func() {
		for _, player := range r.Players {
			player.Conn.WriteJSON(map[string]interface{}{
				"type":    utils.GameActionRoomDestroyed,
				"payload": map[string]any{"room_id": r.ID},
			})
		}

		if r.onRoomDestroyed != nil {
			r.onRoomDestroyed(r.ID)
		}
	})
}

func (r *Room) resolveMatchWinner() *uint {
	if len(r.roundWins) == 0 {
		return nil
	}

	maxWins := -1
	bestPlayer := ""
	tie := false
	for _, playerID := range r.playerOrder {
		wins := r.roundWins[playerID]
		if wins > maxWins {
			maxWins = wins
			bestPlayer = playerID
			tie = false
		} else if wins == maxWins {
			tie = true
		}
	}

	if tie || bestPlayer == "" {
		return nil
	}

	winnerID, err := r.resolvePlayerRecord(bestPlayer)
	if err != nil {
		log.Printf("match end: failed to resolve winner: %v", err)
		return nil
	}
	return &winnerID
}

func (r *Room) updatePlayerStats(winnerID *uint) {
	if r.db == nil {
		return
	}

	for _, playerID := range r.playerOrder {
		recordID, err := r.resolvePlayerRecord(playerID)
		if err != nil {
			log.Printf("match stats: failed to resolve player: %v", err)
			continue
		}
		updates := map[string]any{
			"total_matches": gorm.Expr("total_matches + ?", 1),
		}
		if winnerID != nil {
			if recordID == *winnerID {
				updates["total_wins"] = gorm.Expr("total_wins + ?", 1)
			} else {
				updates["total_losses"] = gorm.Expr("total_losses + ?", 1)
			}
		}
		if err := r.db.Model(&models.Players{}).Where("id = ?", recordID).Updates(updates).Error; err != nil {
			log.Printf("match stats: failed to update player stats: %v", err)
		}
	}
}

func (r *Room) GetPlayer(playerID string) (*Player, bool) {
	player, exists := r.Players[playerID]
	return player, exists
}
