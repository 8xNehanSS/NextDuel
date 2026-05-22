package game

import (
	"fmt"
	"nextduel-svr/models"
	"nextduel-svr/utils"

	"gorm.io/gorm"
)

type GameManager struct {
	rooms       map[string]*Room
	playerRooms map[string]string
	matchmaker  *Matchmaking

	db *gorm.DB
}

func NewGameManager(db *gorm.DB) *GameManager {
	return &GameManager{
		rooms:       make(map[string]*Room),
		playerRooms: make(map[string]string),
		db:          db,
		matchmaker:  NewMatchmaking(),
	}
}

func (gm *GameManager) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"rooms":            len(gm.rooms),
		"queue":            len(gm.matchmaker.GetQueue()),
		"players":          len(gm.matchmaker.GetQueue()) + len(gm.rooms)*2,
		"players_in_rooms": len(gm.rooms) * 2,
		"players_in_queue": len(gm.matchmaker.GetQueue()),
		"memory_usage":     utils.GetMemoryUsage(),
		"cpu_usage":        utils.GetCPUUsage(),
	}
}

func (gm *GameManager) AddToQueue(p *models.Player) (*Room, bool, error) {
	if _, inRoom := gm.playerRooms[p.ID]; inRoom {
		return nil, false, fmt.Errorf("Player already in a room")
	}

	if gm.matchmaker.IsPlayerInQueue(p) {
		return nil, false, fmt.Errorf("Player already in queue")
	}

	room, matched := gm.matchmaker.AddPlayer(p)
	if matched {
		gm.rooms[room.ID] = room
		players := room.GetPlayers()
		for _, player := range players {
			gm.AddPlayerToRoom(player.ID, room.ID)
		}
	}
	return room, matched, nil
}

func (gm *GameManager) GetRoomByPlayerID(playerID string) (*Room, bool) {
	roomID, exists := gm.playerRooms[playerID]
	if !exists {
		return nil, false
	}
	room, exists := gm.rooms[roomID]
	return room, exists
}

func (gm *GameManager) AddPlayerToRoom(playerID string, roomID string) bool {
	if _, roomExists := gm.rooms[roomID]; !roomExists {
		return false
	}
	if _, playerInRoom := gm.playerRooms[playerID]; playerInRoom {
		return false
	}
	gm.playerRooms[playerID] = roomID
	return true
}

func (gm *GameManager) CreateRoom(p *models.Player) *Room {
	room := NewRoom(p, nil)
	gm.rooms[room.ID] = room
	gm.AddPlayerToRoom(p.ID, room.ID)
	return room
}

func (gm *GameManager) GetRoomByID(roomID string) (*Room, bool) {
	room, exists := gm.rooms[roomID]
	return room, exists
}
