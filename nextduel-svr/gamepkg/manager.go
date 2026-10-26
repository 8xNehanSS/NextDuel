package gamepkg

import (
	"nextduel-svr/utils"
	"strings"
	"sync"

	serverutils "nex-server/utils"
)

// GameManager owns websocket sessions, the matchmaking queues (one per game
// mode) and live rooms. All of its maps are guarded by mu; see Room for the
// lock-ordering rule.
type GameManager struct {
	settings Settings
	modes    Modes
	recorder Recorder

	mu           sync.Mutex
	sessions     map[Conn]*Player
	rooms        map[string]*Room
	playerRooms  map[string]string // player ID -> room ID
	queues       map[utils.GameMode]*Matchmaking
	shuttingDown bool
}

func NewGameManager(settings Settings, modes Modes, recorder Recorder) *GameManager {
	if recorder == nil {
		recorder = noopRecorder{}
	}
	queues := make(map[utils.GameMode]*Matchmaking, len(modes))
	for id := range modes {
		queues[id] = NewMatchmaking()
	}
	return &GameManager{
		settings:    settings,
		modes:       modes,
		recorder:    recorder,
		sessions:    make(map[Conn]*Player),
		rooms:       make(map[string]*Room),
		playerRooms: make(map[string]string),
		queues:      queues,
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
	gm.removeFromQueuesLocked(p)
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

// JoinQueue matches p against the next player waiting for the same mode.
func (gm *GameManager) JoinQueue(p *Player, mode utils.GameMode) error {
	mode = normaliseMode(mode)
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if err := gm.checkAvailableLocked(p); err != nil {
		return err
	}
	queue, ok := gm.queues[mode]
	if !ok {
		return ErrUnknownMode
	}

	pair, position := queue.Enqueue(p)
	if pair == nil {
		p.Send(utils.GameActionWaitingQueue, QueuePayload{Mode: mode, Position: position})
		return nil
	}

	room := gm.newRoomLocked(false, mode)
	for _, member := range pair {
		// Cannot fail: the room is brand new and has capacity for two.
		_ = room.AddPlayer(member, utils.GameActionMatchFound)
		gm.playerRooms[member.ID] = room.ID
	}
	return nil
}

func (gm *GameManager) LeaveQueue(p *Player) error {
	gm.mu.Lock()
	removed := gm.removeFromQueuesLocked(p)
	gm.mu.Unlock()
	if !removed {
		return ErrNotQueued
	}
	p.Send(utils.GameActionQueueLeft, nil)
	return nil
}

// ---------------------------------------------------------------------------
// Rooms

func (gm *GameManager) CreateRoom(p *Player, mode utils.GameMode) error {
	mode = normaliseMode(mode)
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if err := gm.checkAvailableLocked(p); err != nil {
		return err
	}
	if _, ok := gm.modes[mode]; !ok {
		return ErrUnknownMode
	}
	room := gm.newRoomLocked(true, mode)
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
	Rooms         int `json:"rooms"`
	ActiveMatches int `json:"active_matches"`
	Queue         int `json:"queue"`
	// Queues is the number of players waiting per game mode.
	Queues         map[utils.GameMode]int `json:"queues"`
	Players        int                    `json:"players"`
	PlayersInRooms int                    `json:"players_in_rooms"`
	PlayersInQueue int                    `json:"players_in_queue"`
	MemoryUsage    string                 `json:"memory_usage"`
	CPUUsage       string                 `json:"cpu_usage"`
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
	queues := make(map[utils.GameMode]int, len(gm.queues))
	queued := 0
	for mode, q := range gm.queues {
		queues[mode] = q.Len()
		queued += q.Len()
	}
	stats := GameStatistics{
		Rooms:          len(gm.rooms),
		ActiveMatches:  active,
		Queue:          queued,
		Queues:         queues,
		Players:        len(online),
		PlayersInRooms: len(gm.playerRooms),
		PlayersInQueue: queued,
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
	var queued []*Player
	for _, q := range gm.queues {
		queued = append(queued, q.Drain()...)
	}
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
	}
	for _, q := range gm.queues {
		if q.Contains(p.ID) {
			return ErrAlreadyQueued
		}
	}
	return nil
}

func (gm *GameManager) removeFromQueuesLocked(p *Player) bool {
	removed := false
	for _, q := range gm.queues {
		removed = q.Remove(p) || removed
	}
	return removed
}

// newRoomLocked requires mode to be registered.
func (gm *GameManager) newRoomLocked(private bool, mode utils.GameMode) *Room {
	id := utils.GenerateRoomID()
	for gm.rooms[id] != nil {
		id = utils.GenerateRoomID()
	}
	room := newRoom(id, private, gm.settings, gm.modes[mode], gm.recorder, gm.onRoomClosed)
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

func normaliseMode(mode utils.GameMode) utils.GameMode {
	if mode == "" {
		return utils.GameModeTrivia
	}
	return mode
}
