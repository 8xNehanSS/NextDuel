import { create } from "zustand";

export type Player = {
    id: string;
    username: string;
    score: number;
};

export type Question = {
    id: string;
    question: string;
    answers: string[];
};

type GamePhase =
    | "none"
    | "waiting"
    | "matched"
    | "started"
    | "game_over"
    | "intermission";

type GameStore = {
    connected: boolean;
    roomId: string | null;
    ready: boolean;

    // Players
    me: Player | null;
    opponent: Player | null;

    // Match state
    phase: GamePhase;
    round: number;
    maxRounds: number;

    // Question
    currentQuestion: Question | null;
    answered: boolean;

    // Timer
    timer: number;

    // Results
    roundWinnerId: string | null;
    gameWinnerId: string | null;

    // Loading
    loading: boolean;

    // Actions
    setConnected: (connected: boolean) => void;

    setReady: (ready: boolean) => void;

    setRoom: (roomId: string) => void;

    setPlayers: (me: Player, opponent: Player) => void;

    setPhase: (phase: GamePhase) => void;

    setQuestion: (question: Question | null) => void;

    setTimer: (timer: number) => void;

    setRound: (round: number) => void;

    setRoundWinner: (playerId: string | null) => void;

    setGameWinner: (playerId: string | null) => void;

    incrementScore: (playerId: string) => void;

    setAnswered: () => void;

    resetGame: () => void;
};

const initialState = {
    connected: false,
    roomId: null,

    ready: false,

    me: null,
    opponent: null,

    phase: "none" as GamePhase,

    round: 1,
    maxRounds: 5,

    currentQuestion: null,
    answered: false,

    timer: 0,

    roundWinnerId: null,
    gameWinnerId: null,

    loading: false,
};

export const useGameStore = create<GameStore>((set) => ({
    ...initialState,

    setReady: (ready) =>
        set({
            ready,
        }),

    setConnected: (connected) =>
        set({
            connected,
        }),

    setRoom: (roomId) =>
        set({
            roomId,
        }),

    setPlayers: (me, opponent) =>
        set({
            me,
            opponent,
        }),

    setPhase: (phase) =>
        set({
            phase,
        }),

    setQuestion: (currentQuestion) =>
        set({
            currentQuestion,
            answered: false, // Reset answered state when a new question is set
        }),

    setTimer: (timer) =>
        set({
            timer,
        }),

    setRound: (round) =>
        set({
            round,
        }),

    setRoundWinner: (roundWinnerId) =>
        set({
            roundWinnerId,
        }),

    setGameWinner: (gameWinnerId) =>
        set({
            gameWinnerId,
        }),
    
    setAnswered: () => 
        set({
            answered: true,
        }),

    incrementScore: (playerId) =>
        set((state) => {
            if (state.me?.id === playerId) {
                return {
                    me: {
                        ...state.me,
                        score: state.me.score + 1,
                    },
                };
            }

            if (state.opponent?.id === playerId) {
                return {
                    opponent: {
                        ...state.opponent,
                        score: state.opponent.score + 1,
                    },
                };
            }

            return state;
        }),

    resetGame: () =>
        set({
            ...initialState,
        }),
}));