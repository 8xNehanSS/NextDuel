// Websocket payloads. Mirrors nextduel-svr/gamepkg/protocol.go — keep in sync.

/** Mirrors utils.GameMode in nextduel-svr/utils/game_enums.go. */
export type GameMode = "trivia" | "math" | "scramble" | "rps";

export type RoomStatus = "waiting" | "starting" | "active" | "completed";

export interface PlayerState {
    id: string;
    ready: boolean;
    score: number;
    roundsWon: number;
}

export interface RoomState {
    roomId: string;
    mode: GameMode;
    private: boolean;
    status: RoomStatus;
    round: number;
    maxRounds: number;
    players: PlayerState[];
}

export interface GameErrorPayload {
    code: string;
    message: string;
}

export interface QueuePayload {
    mode: GameMode;
    position: number;
}

export interface PlayerLeftPayload {
    playerId: string;
}

export interface DuelStartPayload {
    maxRounds: number;
}

export interface RoundStartPayload {
    round: number;
    maxRounds: number;
    question: string;
    category: string;
    difficulty: string;
    choices: string[];
    duration: number;
}

export interface TimerPayload {
    remaining: number;
}

export interface DuelFastPayload {
    playerId: string;
    remaining: number;
}

export interface AnswerAckPayload {
    round: number;
    answer: string;
}

export interface PlayerRoundResult {
    playerId: string;
    answer: string;
    correct: boolean;
    delta: number;
    score: number;
    roundsWon: number;
}

export interface RoundResultPayload {
    round: number;
    /** Empty for modes without a single correct answer (rps). */
    correctAnswer: string;
    summary: string;
    winnerId: string | null;
    results: PlayerRoundResult[];
}

export interface IntermissionPayload {
    remaining: number;
    nextRound: number;
}

export type GameOverReason = "completed" | "forfeit";

export interface GameOverPayload {
    reason: GameOverReason;
    winnerId: string | null;
    players: PlayerState[];
}

export type RoomCloseReason = "opponent_left" | "questions_unavailable" | "server_shutdown";

export interface RoomDestroyedPayload {
    roomId: string;
    reason: RoomCloseReason;
}

export interface ModeRequest {
    mode: GameMode;
}

export interface JoinRoomRequest {
    roomId: string;
}

export interface DuelAnswerRequest {
    round: number;
    answer: string;
}
