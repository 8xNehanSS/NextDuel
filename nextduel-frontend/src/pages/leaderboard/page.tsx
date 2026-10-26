import { useEffect, useState } from "react";
import { getLeaderboard } from "../../services/api/history";
import type { LeaderboardEntry } from "../../types/api";
import { errorMessage } from "../../utils/apiclient";
import { useUserStore } from "../../store/userStore";

const LeaderboardPage = () => {
    const me = useUserStore((s) => s.username);
    const [entries, setEntries] = useState<LeaderboardEntry[] | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const controller = new AbortController();
        getLeaderboard(controller.signal)
            .then(setEntries)
            .catch((err) => {
                if (!controller.signal.aborted) setError(errorMessage(err));
            });
        return () => controller.abort();
    }, []);

    return (
        <div className="page-shell">
            <h1 className="section-title">Leaderboard</h1>
            <p className="section-subtitle">Ranked by wins, then win rate.</p>

            {error && <p className="form-error" style={{ marginTop: 20 }}>Could not load leaderboard: {error}</p>}
            {!entries && !error && <p className="section-subtitle" style={{ marginTop: 20 }}>Loading...</p>}

            {entries && entries.length === 0 && (
                <div className="panel empty-state" style={{ marginTop: 20 }}>
                    <p className="section-subtitle">No ranked matches have been played yet.</p>
                </div>
            )}

            {entries && entries.length > 0 && (
                <div className="panel table-wrap" style={{ marginTop: 20 }}>
                    <table className="table">
                        <thead>
                            <tr>
                                <th>#</th>
                                <th>Player</th>
                                <th className="num">Wins</th>
                                <th className="num">Losses</th>
                                <th className="num">Draws</th>
                                <th className="num">Matches</th>
                                <th className="num">Win rate</th>
                            </tr>
                        </thead>
                        <tbody>
                            {entries.map((e) => (
                                <tr key={e.username} className={e.username === me ? "row-me" : undefined}>
                                    <td>{e.rank}</td>
                                    <td>
                                        {e.username}
                                        {e.username === me && <span className="muted"> (you)</span>}
                                    </td>
                                    <td className="num">{e.wins}</td>
                                    <td className="num">{e.losses}</td>
                                    <td className="num">{e.draws}</td>
                                    <td className="num">{e.total_matches}</td>
                                    <td className="num">{Math.round(e.win_rate * 100)}%</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
};

export default LeaderboardPage;
