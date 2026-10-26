// Mirrors nextduel-svr/utils/events.go — keep in sync.
import type {
    AnswerAckPayload,
    DuelAnswerRequest,
    DuelFastPayload,
    DuelStartPayload,
    GameErrorPayload,
    GameOverPayload,
    IntermissionPayload,
    JoinRoomRequest,
    ModeRequest,
    PlayerLeftPayload,
    QueuePayload,
    RoomDestroyedPayload,
    RoomState,
    RoundResultPayload,
    RoundStartPayload,
    TimerPayload,
} from "../../types/protocol";

export const GameAction = {
    // Server -> client
    ERROR: "error",
    PONG: "pong",
    WAITING_QUEUE: "waiting_queue",
    QUEUE_LEFT: "queue_left",
    MATCH_FOUND: "match_found",
    ROOM_CREATED: "room_created",
    ROOM_JOINED: "room_joined",
    ROOM_STATE: "room_state",
    PLAYER_LEFT: "player_left",
    ROOM_DESTROYED: "room_destroyed",
    DUEL_START: "duel_start",
    ROUND_START: "round_start",
    TIMER_UPDATE: "timer_update",
    DUEL_FAST: "duel_fast",
    ROUND_RESULT: "round_result",
    INTERMISSION: "intermission",
    INTERMISSION_TIMER_UPDATE: "intermission_timer_update",
    GAME_OVER: "game_over",

    // Client -> server
    PING: "ping",
    JOIN_QUEUE: "join_queue",
    LEAVE_QUEUE: "leave_queue",
    CREATE_ROOM: "create_room",
    JOIN_ROOM: "join_room",
    LEAVE_ROOM: "leave_room",
    PLAYER_READY: "player_ready",
    DUEL_ANSWER: "duel_answer",
} as const;

export type GameAction = (typeof GameAction)[keyof typeof GameAction];

/** Payload type of each server -> client message. */
export interface ServerEvents {
    [GameAction.ERROR]: GameErrorPayload;
    [GameAction.PONG]: null;
    [GameAction.WAITING_QUEUE]: QueuePayload;
    [GameAction.QUEUE_LEFT]: null;
    [GameAction.MATCH_FOUND]: RoomState;
    [GameAction.ROOM_CREATED]: RoomState;
    [GameAction.ROOM_JOINED]: RoomState;
    [GameAction.ROOM_STATE]: RoomState;
    [GameAction.PLAYER_LEFT]: PlayerLeftPayload;
    [GameAction.ROOM_DESTROYED]: RoomDestroyedPayload;
    [GameAction.DUEL_START]: DuelStartPayload;
    [GameAction.ROUND_START]: RoundStartPayload;
    [GameAction.TIMER_UPDATE]: TimerPayload;
    [GameAction.DUEL_FAST]: DuelFastPayload;
    [GameAction.DUEL_ANSWER]: AnswerAckPayload;
    [GameAction.ROUND_RESULT]: RoundResultPayload;
    [GameAction.INTERMISSION]: IntermissionPayload;
    [GameAction.INTERMISSION_TIMER_UPDATE]: TimerPayload;
    [GameAction.GAME_OVER]: GameOverPayload;
}

/** Payload type of each client -> server message. */
export interface ClientEvents {
    [GameAction.PING]: undefined;
    [GameAction.JOIN_QUEUE]: ModeRequest;
    [GameAction.LEAVE_QUEUE]: undefined;
    [GameAction.CREATE_ROOM]: ModeRequest;
    [GameAction.JOIN_ROOM]: JoinRoomRequest;
    [GameAction.LEAVE_ROOM]: undefined;
    [GameAction.PLAYER_READY]: undefined;
    [GameAction.DUEL_ANSWER]: DuelAnswerRequest;
}

export type ServerEventType = keyof ServerEvents;
export type ClientEventType = keyof ClientEvents;
