import { api } from "../../utils/apiclient";
import type { DuelHistory, LeaderboardEntry } from "../../types/api";

export function getDuelHistory(limit: number, offset: number, signal?: AbortSignal): Promise<DuelHistory> {
    return api.get<DuelHistory>("/history/duels", { query: { limit, offset }, signal });
}

export function getLeaderboard(signal?: AbortSignal): Promise<LeaderboardEntry[]> {
    return api.get<LeaderboardEntry[]>("/leaderboard", { signal });
}
