package gamepkg

import (
	"fmt"
	"log"
	serverutils "nex-server/utils"
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
		matchmaker:  NewMatchmaking(db),
	}
}

type GameStatistics struct {
	Rooms          int    `json:"rooms"`
	Queue          int    `json:"queue"`
	Players        int    `json:"players"`
	PlayersInRooms int    `json:"players_in_rooms"`
	PlayersInQueue int    `json:"players_in_queue"`
	MemoryUsage    string `json:"memory_usage"`
	CPUUsage       string `json:"cpu_usage"`
}

func (gm *GameManager) GetStatistics() GameStatistics {
	return GameStatistics{
		Rooms:          len(gm.rooms),
		Queue:          len(gm.matchmaker.GetQueue()),
		Players:        len(gm.playerRooms),
		PlayersInRooms: len(gm.playerRooms) - len(gm.matchmaker.GetQueue()),
		PlayersInQueue: len(gm.matchmaker.GetQueue()),
		MemoryUsage:    serverutils.GetMemoryUsage(),
		CPUUsage:       serverutils.GetCPUUsage(),
	}
}

func (gm *GameManager) AddToQueue(p *Player) (*Room, bool, error) {
	if _, inRoom := gm.playerRooms[p.ID]; inRoom {
		return nil, false, fmt.Errorf("Player already in a room")
	}

	if gm.matchmaker.IsPlayerInQueue(p) {
		return nil, false, fmt.Errorf("Player already in queue")
	}

	room, matched := gm.matchmaker.AddPlayer(p)
	if matched {
		gm.attachRoomLifecycle(room)
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

func (gm *GameManager) CreateRoom(p *Player) *Room {
	room := NewRoom(p, nil, gm.db)
	gm.attachRoomLifecycle(room)
	gm.rooms[room.ID] = room
	gm.AddPlayerToRoom(p.ID, room.ID)
	return room
}

func (gm *GameManager) GetRoomByID(roomID string) (*Room, bool) {
	room, exists := gm.rooms[roomID]
	return room, exists
}

func (gm *GameManager) HandlePlayerDisconnect(player *Player) {
	if player == nil {
		return
	}

	gm.matchmaker.RemovePlayer(player.ID)

	room, inRoom := gm.GetRoomByPlayerID(player.ID)
	if !inRoom {
		return
	}

	room.HandleDisconnect(player.ID)

	delete(gm.playerRooms, player.ID)
	delete(room.Players, player.ID)
	delete(room.Scores, player.ID)

	opponent := room.GetOpponent(player.ID)
	if opponent != nil {
		if opponent.Conn != nil {
			opponent.Conn.WriteJSON(map[string]any{
				"type":    utils.GameActionPlayerLeft,
				"payload": player.ID,
			})
			opponent.Conn.WriteJSON(map[string]any{
				"type": utils.GameActionRoomDestroyed,
				"payload": map[string]any{
					"room_id":   room.ID,
					"player_id": player.ID,
				},
			})
		}

		select {
		case room.stopTimer <- struct{}{}:
		default:
		}

		delete(gm.playerRooms, opponent.ID)
	}

	delete(gm.rooms, room.ID)
}

func (gm *GameManager) DestroyRoom(roomID string) error {
	room, exists := gm.rooms[roomID]
	if !exists {
		return fmt.Errorf("Room not found")
	}
	for playerID := range room.Players {
		delete(gm.playerRooms, playerID)
	}
	delete(gm.rooms, roomID)
	return nil
}

func (gm *GameManager) attachRoomLifecycle(room *Room) {
	if room == nil {
		return
	}

	room.onRoomDestroyed = func(roomID string) {
		if err := gm.DestroyRoom(roomID); err != nil {
			log.Printf("room cleanup: failed to destroy room %s: %v", roomID, err)
		}
	}
}
