package gamepkg

import (
	"nextduel-svr/questions"
	"nextduel-svr/utils"
	"strings"
	"sync"

	serverutils "nex-server/utils"
)

// GameManager owns websocket sessions, the matchmaking queue and live rooms.
// All of its maps are guarded by mu; see Room for the lock-ordering rule.
type GameManager struct {
	settings Settings
	provider questions.Provider
	recorder Recorder

	mu           sync.Mutex
	sessions     map[Conn]*Player
	rooms        map[string]*Room
	playerRooms  map[string]string // player ID -> room ID
	queue        *Matchmaking
	shuttingDown bool
}

func NewGameManager(settings Settings, provider questions.Provider, recorder Recorder) *GameManager {
	if recorder == nil {
		recorder = noopRecorder{}
	}
	return &GameManager{
		settings:    settings,
		provider:    provider,
		recorder:    recorder,
		sessions:    make(map[Conn]*Player),
		rooms:       make(map[string]*Room),
		playerRooms: make(map[string]string),
		queue:       NewMatchmaking(),
	}
}

// ---------------------------------------------------------------------------
// Sessions

// Session returns the Player for conn, creating it on first use. A user with
// two tabs open has two sessions; only the session that joined a room can act
// in it.
func (gm *GameManager) Session(userID string, conn Conn) *Player {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	if p, ok := gm.sessions[conn]; ok {
		return p
	}
	p := NewPlayer(userID, conn)
	gm.sessions[conn] = p
	return p
}

// EndSession is called when a websocket closes. Leaving mid-match forfeits.
func (gm *GameManager) EndSession(conn Conn) {
	gm.mu.Lock()
	p, ok := gm.sessions[conn]
	if !ok {
		gm.mu.Unlock()
		return
	}
	delete(gm.sessions, conn)
	gm.queue.Remove(p)
	room := gm.roomOfLocked(p.ID)
	gm.mu.Unlock()

	if room != nil {
		if err := room.Leave(p); err == nil {
			gm.unmapPlayer(p.ID, room.ID)
		}
	}
	p.Close()
}

// ---------------------------------------------------------------------------
// Queue

func (gm *GameManager) JoinQueue(p *Player) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if err := gm.checkAvailableLocked(p); err != nil {
		return err
	}

	pair, position := gm.queue.Enqueue(p)
	if pair == nil {
		p.Send(utils.GameActionWaitingQueue, QueuePayload{Position: position})
		return nil
	}

	room := gm.newRoomLocked(false)
	for _, member := range pair {
		// Cannot fail: the room is brand new and has capacity for two.
		_ = room.AddPlayer(member, utils.GameActionMatchFound)
		gm.playerRooms[member.ID] = room.ID
	}
	return nil
}

func (gm *GameManager) LeaveQueue(p *Player) error {
	gm.mu.Lock()
	removed := gm.queue.Remove(p)
	gm.mu.Unlock()
	if !removed {
		return ErrNotQueued
	}
	p.Send(utils.GameActionQueueLeft, nil)
	return nil
}

// ---------------------------------------------------------------------------
// Rooms

func (gm *GameManager) CreateRoom(p *Player) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if err := gm.checkAvailableLocked(p); err != nil {
		return err
	}
	room := gm.newRoomLocked(true)
	if err := room.AddPlayer(p, utils.GameActionRoomCreated); err != nil {
		delete(gm.rooms, room.ID)
		return err
	}
	gm.playerRooms[p.ID] = room.ID
	return nil
}

func (gm *GameManager) JoinRoom(p *Player, code string) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return ErrBadRequest
	}

	gm.mu.Lock()
	defer gm.mu.Unlock()

	if err := gm.checkAvailableLocked(p); err != nil {
		return err
	}
	room, ok := gm.rooms[code]
	if !ok || !room.Private {
		return ErrRoomNotFound
	}
	if err := room.AddPlayer(p, utils.GameActionRoomJoined); err != nil {
		return err
	}
	gm.playerRooms[p.ID] = room.ID
	return nil
}

func (gm *GameManager) LeaveRoom(p *Player) error {
	room := gm.RoomOf(p.ID)
	if room == nil {
		return ErrNotInRoom
	}
	if err := room.Leave(p); err != nil {
		return err
	}
	gm.unmapPlayer(p.ID, room.ID)
	return nil
}

func (gm *GameManager) SetReady(p *Player) error {
	room := gm.RoomOf(p.ID)
	if room == nil {
		return ErrNotInRoom
	}
	return room.SetReady(p)
}

func (gm *GameManager) SubmitAnswer(p *Player, round int, answer string) error {
	room := gm.RoomOf(p.ID)
	if room == nil {
		return ErrNotInRoom
	}
	return room.SubmitAnswer(p, round, answer)
}

func (gm *GameManager) RoomOf(playerID string) *Room {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	return gm.roomOfLocked(playerID)
}

// ---------------------------------------------------------------------------
// Lifecycle

type GameStatistics struct {
	Rooms          int    `json:"rooms"`
	ActiveMatches  int    `json:"active_matches"`
	Queue          int    `json:"queue"`
	Players        int    `json:"players"`
	PlayersInRooms int    `json:"players_in_rooms"`
	PlayersInQueue int    `json:"players_in_queue"`
	MemoryUsage    string `json:"memory_usage"`
	CPUUsage       string `json:"cpu_usage"`
}

func (gm *GameManager) Statistics() GameStatistics {
	gm.mu.Lock()
	online := make(map[string]struct{}, len(gm.sessions))
	for _, p := range gm.sessions {
		online[p.ID] = struct{}{}
	}
	active := 0
	for _, room := range gm.rooms {
		if room.Status() == utils.RoomStatusActive {
			active++
		}
	}
	stats := GameStatistics{
		Rooms:          len(gm.rooms),
		ActiveMatches:  active,
		Queue:          gm.queue.Len(),
		Players:        len(online),
		PlayersInRooms: len(gm.playerRooms),
		PlayersInQueue: gm.queue.Len(),
	}
	gm.mu.Unlock()

	// Sampling CPU can take a moment; do it outside the lock.
	stats.MemoryUsage = serverutils.GetMemoryUsage()
	stats.CPUUsage = serverutils.GetCPUUsage()
	return stats
}

// Shutdown refuses new games, aborts live rooms and flushes every session.
func (gm *GameManager) Shutdown() {
	gm.mu.Lock()
	gm.shuttingDown = true
	rooms := make([]*Room, 0, len(gm.rooms))
	for _, room := range gm.rooms {
		rooms = append(rooms, room)
	}
	queued := gm.queue.Drain()
	gm.mu.Unlock()

	for _, room := range rooms {
		room.Abort(utils.RoomCloseServerShutdown)
	}
	for _, p := range queued {
		p.SendError(ErrShuttingDown)
	}

	gm.mu.Lock()
	sessions := make([]*Player, 0, len(gm.sessions))
	for _, p := range gm.sessions {
		sessions = append(sessions, p)
	}
	gm.mu.Unlock()
	for _, p := range sessions {
		p.Close()
	}
}

// ---------------------------------------------------------------------------
// Internals

func (gm *GameManager) checkAvailableLocked(p *Player) error {
	switch {
	case gm.shuttingDown:
		return ErrShuttingDown
	case gm.playerRooms[p.ID] != "":
		return ErrAlreadyInRoom
	case gm.queue.Contains(p.ID):
		return ErrAlreadyQueued
	}
	return nil
}

func (gm *GameManager) newRoomLocked(private bool) *Room {
	id := utils.GenerateRoomID()
	for gm.rooms[id] != nil {
		id = utils.GenerateRoomID()
	}
	room := newRoom(id, private, gm.settings, gm.provider, gm.recorder, gm.onRoomClosed)
	gm.rooms[id] = room
	return room
}

func (gm *GameManager) roomOfLocked(playerID string) *Room {
	id, ok := gm.playerRooms[playerID]
	if !ok {
		return nil
	}
	return gm.rooms[id]
}

func (gm *GameManager) onRoomClosed(roomID string, playerIDs []string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	delete(gm.rooms, roomID)
	for _, id := range playerIDs {
		if gm.playerRooms[id] == roomID {
			delete(gm.playerRooms, id)
		}
	}
}

func (gm *GameManager) unmapPlayer(playerID, roomID string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	if gm.playerRooms[playerID] == roomID {
		delete(gm.playerRooms, playerID)
	}
}
