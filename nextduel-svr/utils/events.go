package utils

type GameAction string

const (
	GameActionConnect    GameAction = "connect"
	GameActionDisconnect GameAction = "disconnect"
	GameActionError      GameAction = "error"

	GameActionMessage GameAction = "message"
	GameActionTyping  GameAction = "typing"

	GameActionPing GameAction = "ping"
	GameActionPong GameAction = "pong"

	GameActionJoinQueue    GameAction = "join_queue"
	GameActionPlayerReady  GameAction = "player_ready"
	GameActionCreateRoom   GameAction = "create_room"
	GameActionJoinRoom     GameAction = "join_room"
	GameActionPlayerJoined GameAction = "player_joined"

	GameActionDuelAnswer   GameAction = "duel_answer"
	GameActionDuelStart    GameAction = "duel_start"
	GameActionDuelFast     GameAction = "duel_fast"
	GameActionDuelComplete GameAction = "duel_complete"

	GameActionGameOver GameAction = "game_over"
)
