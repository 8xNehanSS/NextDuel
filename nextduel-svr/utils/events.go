package utils

type GameAction string

const (
	// Connection
	GameActionConnected    GameAction = "connected"
	GameActionDisconnected GameAction = "disconnected"
	GameActionError        GameAction = "error"
	GameActionPong         GameAction = "pong"

	// Queue
	GameActionWaitingQueue GameAction = "waiting_queue"
	GameActionMatchFound   GameAction = "match_found"
	GameActionQueueCount   GameAction = "queue_count"

	// Room
	GameActionRoomCreated   GameAction = "room_created"
	GameActionRoomJoined    GameAction = "room_joined"
	GameActionPlayerJoined  GameAction = "player_joined"
	GameActionPlayerLeft    GameAction = "player_left"
	GameActionRoomDestroyed GameAction = "room_destroyed"

	// Match Lifecycle
	GameActionDuelStart    GameAction = "duel_start"
	GameActionRoundStart   GameAction = "round_start"
	GameActionRoundResult  GameAction = "round_result"
	GameActionDuelFast     GameAction = "duel_fast"
	GameActionDuelComplete GameAction = "duel_complete"
	GameActionIntermission GameAction = "intermission"

	// Game State
	GameActionScoreUpdate             GameAction = "score_update"
	GameActionTimerUpdate             GameAction = "timer_update"
	GameActionIntermissionTimerUpdate GameAction = "intermission_timer_update"
	GameActionGameOver                GameAction = "game_over"

	// Presence
	GameActionPlayerTyping GameAction = "player_typing"

	// Chat
	GameActionMessageReceived GameAction = "message_received"
)

const (
	// Connection
	GameActionConnect GameAction = "connect"
	GameActionPing    GameAction = "ping"

	// Queue
	GameActionJoinQueue  GameAction = "join_queue"
	GameActionLeaveQueue GameAction = "leave_queue"

	// Room
	GameActionCreateRoom GameAction = "create_room"
	GameActionJoinRoom   GameAction = "join_room"

	// Match
	GameActionPlayerReady   GameAction = "player_ready"
	GameActionPlayerReadyOk GameAction = "player_ready_ok"

	// Gameplay
	GameActionDuelAnswer GameAction = "duel_answer"
	GameActionTyping     GameAction = "typing"

	// Chat
	GameActionMessage GameAction = "message"
)
