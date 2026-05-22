export const GameAction = {
  CONNECT: "connect",
  DISCONNECT: "disconnect",
  ERROR: "error",

  MESSAGE: "message",
  TYPING: "typing",

  PING: "ping",
  PONG: "pong",

  JOIN_QUEUE: "join_queue",
  PLAYER_READY: "player_ready",
  CREATE_ROOM: "create_room",
  JOIN_ROOM: "join_room",
  PLAYER_JOINED: "player_joined",

  DUEL_ANSWER: "duel_answer",
  DUEL_START: "duel_start",
  DUEL_FAST: "duel_fast",
  DUEL_COMPLETE: "duel_complete",

  GAME_OVER: "game_over",
} as const

export type GameAction = typeof GameAction[keyof typeof GameAction]