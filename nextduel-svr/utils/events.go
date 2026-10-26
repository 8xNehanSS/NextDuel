package utils

// GameAction is the "type" field of every websocket message exchanged between
// the client and the server. The TypeScript mirror lives in
// nextduel-frontend/src/utils/websocket/events.ts — keep both in sync.
type GameAction string

// Server -> client
const (
	GameActionError GameAction = "error"
	GameActionPong  GameAction = "pong"

	// Queue
	GameActionWaitingQueue GameAction = "waiting_queue"
	GameActionQueueLeft    GameAction = "queue_left"
	GameActionMatchFound   GameAction = "match_found"

	// Room
	GameActionRoomCreated   GameAction = "room_created"
	GameActionRoomJoined    GameAction = "room_joined"
	GameActionRoomState     GameAction = "room_state"
	GameActionPlayerLeft    GameAction = "player_left"
	GameActionRoomDestroyed GameAction = "room_destroyed"

	// Match lifecycle
	GameActionDuelStart               GameAction = "duel_start"
	GameActionRoundStart              GameAction = "round_start"
	GameActionTimerUpdate             GameAction = "timer_update"
	GameActionDuelFast                GameAction = "duel_fast"
	GameActionRoundResult             GameAction = "round_result"
	GameActionIntermission            GameAction = "intermission"
	GameActionIntermissionTimerUpdate GameAction = "intermission_timer_update"
	GameActionGameOver                GameAction = "game_over"
)

// Client -> server
const (
	GameActionPing GameAction = "ping"

	// Queue
	GameActionJoinQueue  GameAction = "join_queue"
	GameActionLeaveQueue GameAction = "leave_queue"

	// Room
	GameActionCreateRoom GameAction = "create_room"
	GameActionJoinRoom   GameAction = "join_room"
	GameActionLeaveRoom  GameAction = "leave_room"

	// Match
	GameActionPlayerReady GameAction = "player_ready"

	// Gameplay (also echoed back to the sender as an acknowledgement)
	GameActionDuelAnswer GameAction = "duel_answer"
)
