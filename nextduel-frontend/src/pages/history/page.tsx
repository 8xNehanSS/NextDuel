import { useEffect, useState } from "react";
import { Link } from "react-router";
import { getDuelHistory } from "../../services/api/history";
import type { DuelHistory, MatchResult } from "../../types/api";
import { errorMessage } from "../../utils/apiclient";
import type { GameMode } from "../../types/protocol";
import GameTabs from "../../components/GameTabs";
import { gameInfo } from "../../games";

const PAGE_SIZE = 10;

const resultLabel: Record<MatchResult, string> = { win: "Win", loss: "Loss", draw: "Draw" };

function formatDate(iso: string) {
    const d = new Date(iso);
    return Number.isNaN(d.getTime()) || d.getFullYear() < 2000 ? "—" : d.toLocaleString();
}

const HistoryPage = () => {
    const [page, setPage] = useState(0);
    const [mode, setMode] = useState<GameMode | undefined>(undefined);
    const [history, setHistory] = useState<DuelHistory | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const controller = new AbortController();
        getDuelHistory(PAGE_SIZE, page * PAGE_SIZE, mode, controller.signal)
            .then((h) => {
                setHistory(h);
                setError(null);
            })
            .catch((err) => {
                if (!controller.signal.aborted) setError(errorMessage(err));
            });
        return () => controller.abort();
    }, [page, mode]);

    const totalPages = history ? Math.max(1, Math.ceil(history.total / PAGE_SIZE)) : 1;
    const stats = history?.stats;

    return (
        <div className="page-shell">
            <h1 className="section-title">Match history</h1>
            <p className="section-subtitle">Your finished duels, newest first.</p>

            <GameTabs
                value={mode}
                onChange={(m) => {
                    setMode(m);
                    setPage(0);
                }}
            />

            {stats && (
                <div className="stat-grid" style={{ marginTop: 20 }}>
                    <div className="stat-card">
                        <p className="stat-label">Matches</p>
                        <p className="stat-value">{stats.total_matches}</p>
                    </div>
                    <div className="stat-card">
                        <p className="stat-label">Wins</p>
                        <p className="stat-value">{stats.wins}</p>
                    </div>
                    <div className="stat-card">
                        <p className="stat-label">Losses</p>
                        <p className="stat-value">{stats.losses}</p>
                    </div>
                    <div className="stat-card">
                        <p className="stat-label">Draws</p>
                        <p className="stat-value">{stats.draws}</p>
                    </div>
                    <div className="stat-card">
                        <p className="stat-label">Win rate</p>
                        <p className="stat-value">{Math.round(stats.win_rate * 100)}%</p>
                    </div>
                </div>
            )}

            {error && <p className="form-error" style={{ marginTop: 20 }}>Could not load history: {error}</p>}
            {!history && !error && <p className="section-subtitle" style={{ marginTop: 20 }}>Loading...</p>}

            {history && history.matches.length === 0 && (
                <div className="panel empty-state" style={{ marginTop: 20 }}>
                    <h2 className="section-title" style={{ fontSize: 22 }}>
                        No matches yet
                    </h2>
                    <p className="section-subtitle">
                        {mode ? `You haven't finished a ${gameInfo(mode).name} duel yet.` : "Finish a duel and it will show up here."}
                    </p>
                    <div className="hero-actions" style={{ justifyContent: "center" }}>
                        <Link to={mode ? `/lobby?game=${mode}` : "/lobby"} className="btn btn-primary">
                            Play {mode ? gameInfo(mode).name : "your first duel"}
                        </Link>
                    </div>
                </div>
            )}

            {history && history.matches.length > 0 && (
                <>
                    <div className="match-list" style={{ marginTop: 20 }}>
                        {history.matches.map((m) => (
                            <div key={m.match_id} className="panel match-row">
                                <span className={`badge badge-${m.result}`}>{resultLabel[m.result]}</span>
                                <div className="match-main">
                                    <strong>vs {m.opponent || "unknown"}</strong>
                                    <span className="muted">
                                        {gameInfo(m.mode).icon} {gameInfo(m.mode).name}
                                    </span>
                                    <span className="muted">
                                        {formatDate(m.started_at)}
                                        {m.status === "abandoned" && " · ended by forfeit"}
                                    </span>
                                </div>
                                <div className="match-score">
                                    <strong>
                                        {m.rounds_won} – {m.opponent_rounds_won}
                                    </strong>
                                    <span className="muted">
                                        rounds{m.mode !== "rps" && ` · ${m.score}/${m.best_of} correct`}
                                    </span>
                                </div>
                            </div>
                        ))}
                    </div>

                    {totalPages > 1 && (
                        <div className="pagination">
                            <button className="btn" disabled={page === 0} onClick={() => setPage((p) => p - 1)}>
                                Previous
                            </button>
                            <span className="muted">
                                Page {page + 1} of {totalPages}
                            </span>
                            <button className="btn" disabled={page + 1 >= totalPages} onClick={() => setPage((p) => p + 1)}>
                                Next
                            </button>
                        </div>
                    )}
                </>
            )}
        </div>
    );
};

export default HistoryPage;
