import { create } from "zustand";
import type {
    GameOverPayload,
    PlayerState,
    RoomState,
    RoundResultPayload,
    RoundStartPayload,
} from "../types/protocol";

export type GamePhase =
    | "idle" // not queued, not in a room
    | "queued" // waiting in matchmaking
    | "lobby" // in a room, waiting for an opponent and/or ready-ups
    | "starting" // everyone ready, questions loading
    | "question" // a round is live
    | "round_result" // showing the result between rounds
    | "game_over";

/** Phases in which losing the connection loses the game. */
export const IN_GAME_PHASES: ReadonlySet<GamePhase> = new Set(["queued", "lobby", "starting", "question", "round_result"]);

type GameState = {
    connected: boolean;
    phase: GamePhase;
    queuePosition: number;
    room: RoomState | null;
    maxRounds: number;
    question: RoundStartPayload | null;
    timer: number;
    /** Answer the server has acknowledged for the current round. */
    answer: string | null;
    /** Answer sent but not yet acknowledged. */
    pendingAnswer: string | null;
    opponentAnswered: boolean;
    lastResult: RoundResultPayload | null;
    nextRound: number | null;
    gameOver: GameOverPayload | null;
};

type GameActions = {
    setConnected: (connected: boolean) => void;
    queued: (position: number) => void;
    enteredRoom: (room: RoomState) => void;
    roomUpdated: (room: RoomState) => void;
    starting: (maxRounds: number) => void;
    roundStarted: (q: RoundStartPayload) => void;
    tick: (remaining: number) => void;
    answerPending: (answer: string) => void;
    answerRejected: () => void;
    answerAccepted: (answer: string) => void;
    opponentAnsweredFast: (remaining: number) => void;
    roundResolved: (result: RoundResultPayload) => void;
    intermission: (nextRound: number, remaining: number) => void;
    finished: (result: GameOverPayload) => void;
    reset: () => void;
};

const initialState: GameState = {
    connected: false,
    phase: "idle",
    queuePosition: 0,
    room: null,
    maxRounds: 5,
    question: null,
    timer: 0,
    answer: null,
    pendingAnswer: null,
    opponentAnswered: false,
    lastResult: null,
    nextRound: null,
    gameOver: null,
};

function applyResults(room: RoomState | null, result: RoundResultPayload): RoomState | null {
    if (!room) return room;
    const byId = new Map(result.results.map((r) => [r.playerId, r]));
    return {
        ...room,
        round: result.round,
        players: room.players.map((p): PlayerState => {
            const r = byId.get(p.id);
            return r ? { ...p, score: r.score, roundsWon: r.roundsWon } : p;
        }),
    };
}

export const useGameStore = create<GameState & GameActions>((set) => ({
    ...initialState,

    setConnected: (connected) => set({ connected }),

    queued: (queuePosition) => set({ phase: "queued", queuePosition }),

    enteredRoom: (room) =>
        set({ ...initialState, connected: true, phase: "lobby", room, maxRounds: room.maxRounds }),

    roomUpdated: (room) =>
        set((s) => ({
            room,
            maxRounds: room.maxRounds,
            phase: room.status === "starting" && s.phase === "lobby" ? "starting" : s.phase,
        })),

    starting: (maxRounds) => set({ phase: "starting", maxRounds }),

    roundStarted: (question) =>
        set((s) => ({
            phase: "question",
            question,
            maxRounds: question.maxRounds,
            timer: question.duration,
            answer: null,
            pendingAnswer: null,
            opponentAnswered: false,
            nextRound: null,
            room: s.room ? { ...s.room, status: "active", round: question.round } : s.room,
        })),

    tick: (timer) => set({ timer }),

    answerPending: (pendingAnswer) => set({ pendingAnswer }),

    answerRejected: () => set({ pendingAnswer: null }),

    answerAccepted: (answer) => set({ answer, pendingAnswer: null }),

    opponentAnsweredFast: (timer) => set({ opponentAnswered: true, timer }),

    roundResolved: (lastResult) =>
        set((s) => ({ phase: "round_result", lastResult, room: applyResults(s.room, lastResult) })),

    intermission: (nextRound, timer) => set({ nextRound, timer }),

    finished: (gameOver) =>
        set((s) => ({
            phase: "game_over",
            gameOver,
            room: s.room ? { ...s.room, status: "completed", players: gameOver.players } : s.room,
        })),

    reset: () => set((s) => ({ ...initialState, connected: s.connected })),
}));
