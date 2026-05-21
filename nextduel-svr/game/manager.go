package game

import (
	"nextduel-svr/models"
	"nextduel-svr/utils"
)

type GameManager struct {
	Rooms map[string]*Room
	Queue []*models.Player
}

func (gm *GameManager) AddToQueue(p *models.Player) {
	gm.Queue = append(gm.Queue, p)

	if len(gm.Queue) >= 2 {
		p1 := gm.Queue[0]
		p2 := gm.Queue[1]
		gm.Queue = gm.Queue[2:]

		room := gm.createRoom(p1, p2)
		gm.Rooms[room.ID] = room

		room.startGame()
	}
}

func (gm *GameManager) createRoom(p1, p2 *models.Player) *Room {
	roomID := utils.GenerateRoomID()
	return &Room{
		ID:      roomID,
		Players: map[string]*models.Player{p1.ID: p1, p2.ID: p2},
		Scores:  map[string]int{p1.ID: 0, p2.ID: 0},
		Quest:   getRandomQuest(),
		Status:  "waiting",
	}
}

func getRandomQuest() models.Quest {
	return models.Quest{
		ID:       "quest1",
		Question: "What is the capital of France?",
		Answer:   "Paris",
	}
}
