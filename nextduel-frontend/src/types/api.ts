// REST response shapes. Mirrors nextduel-svr/repository and routes.

import type { GameMode } from "./protocol";

export interface User {
    id: number;
    username: string;
    email: string;
    role: string;
    is_active: boolean;
    created_at: string;
}

export interface ServerStatistics {
    rooms: number;
    queues: Partial<Record<GameMode, number>>;
    active_matches: number;
    queue: number;
    players: number;
    players_in_rooms: number;
    players_in_queue: number;
    memory_usage: string;
    cpu_usage: string;
}

export interface PlayerStats {
    total_matches: number;
    wins: number;
    losses: number;
    draws: number;
    win_rate: number;
}

export type MatchResult = "win" | "loss" | "draw";

export interface MatchSummary {
    match_id: number;
    mode: GameMode;
    status: "completed" | "abandoned";
    result: MatchResult;
    best_of: number;
    score: number;
    rounds_won: number;
    opponent: string;
    opponent_score: number;
    opponent_rounds_won: number;
    started_at: string;
    ended_at: string;
}

export interface DuelHistory {
    stats: PlayerStats;
    matches: MatchSummary[];
    total: number;
    limit: number;
    offset: number;
}

export interface LeaderboardEntry extends PlayerStats {
    rank: number;
    username: string;
}
