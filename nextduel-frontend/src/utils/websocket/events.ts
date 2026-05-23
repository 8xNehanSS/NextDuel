export const GameAction = {
  // Connection
  CONNECT: "connect",
  CONNECTED: "connected",
  DISCONNECT: "disconnect",
  DISCONNECTED: "disconnected",
  ERROR: "error",

  // Messaging
  MESSAGE: "message",
  MESSAGE_RECEIVED: "message_received",
  TYPING: "typing",
  PLAYER_TYPING: "player_typing",

  // Heartbeat
  PING: "ping",
  PONG: "pong",

  // Queue
  JOIN_QUEUE: "join_queue",
  LEAVE_QUEUE: "leave_queue",
  WAITING_QUEUE: "waiting_queue",
  MATCH_FOUND: "match_found",

  // Room
  CREATE_ROOM: "create_room",
  ROOM_CREATED: "room_created",

  ROOM_DESTROYED: "room_destroyed",

  JOIN_ROOM: "join_room",
  ROOM_JOINED: "room_joined",

  PLAYER_JOINED: "player_joined",
  PLAYER_LEFT: "player_left",

  // Match Lifecycle
  PLAYER_READY: "player_ready",

  DUEL_START: "duel_start",
  ROUND_START: "round_start",
  ROUND_RESULT: "round_result",

  INTERMISSION: "intermission",

  // Gameplay
  DUEL_ANSWER: "duel_answer",
  DUEL_FAST: "duel_fast",

  SCORE_UPDATE: "score_update",
  TIMER_UPDATE: "timer_update",
  INTERMISSION_TIMER_UPDATE: "intermission_timer_update",
  // End
  DUEL_COMPLETE: "duel_complete",
  GAME_OVER: "game_over",
} as const;

export type GameAction =
  typeof GameAction[keyof typeof GameAction];