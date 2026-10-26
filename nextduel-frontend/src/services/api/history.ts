import { api } from "../../utils/apiclient";
import type { DuelHistory, LeaderboardEntry } from "../../types/api";
import type { GameMode } from "../../types/protocol";

/** mode undefined = all games. */
export function getDuelHistory(limit: number, offset: number, mode?: GameMode, signal?: AbortSignal): Promise<DuelHistory> {
    return api.get<DuelHistory>("/history/duels", { query: { limit, offset, mode }, signal });
}

/** mode undefined = lifetime totals across all games. */
export function getLeaderboard(mode?: GameMode, signal?: AbortSignal): Promise<LeaderboardEntry[]> {
    return api.get<LeaderboardEntry[]>("/leaderboard", { query: { mode }, signal });
}
