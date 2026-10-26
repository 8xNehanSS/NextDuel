import { useEffect, useState } from "react";
import { Link } from "react-router";
import { fetchServerStatistics } from "../../services/api/common";
import type { ServerStatistics } from "../../types/api";
import { errorMessage } from "../../utils/apiclient";
import { useUserStore } from "../../store/userStore";
import { GAMES } from "../../games";

const REFRESH_MS = 10_000;

const StatCard = ({ label, value }: { label: string; value: string | number }) => (
    <div className="stat-card">
        <p className="stat-label">{label}</p>
        <p className="stat-value">{value}</p>
    </div>
);

const Homepage = () => {
    const username = useUserStore((s) => s.username);
    const [stats, setStats] = useState<ServerStatistics | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const controller = new AbortController();
        const load = () =>
            fetchServerStatistics(controller.signal)
                .then((s) => {
                    setStats(s);
                    setError(null);
                })
                .catch((err) => {
                    if (!controller.signal.aborted) setError(errorMessage(err));
                });

        load();
        const timer = window.setInterval(load, REFRESH_MS);
        return () => {
            controller.abort();
            window.clearInterval(timer);
        };
    }, []);

    return (
        <div className="page-shell">
            <div className="hero">
                <div className="panel hero-card float-in">
                    <h1 className="hero-title">Ready for a duel, {username}?</h1>
                    <p className="section-subtitle">
                        Real-time head-to-head duels: trivia, math, words and more. Answer first to put your rival on the clock.
                    </p>
                    <div className="hero-actions">
                        <Link to="/lobby" className="btn btn-primary">
                            Play now
                        </Link>
                        <Link to="/leaderboard" className="btn">
                            View leaderboard
                        </Link>
                    </div>
                </div>

                <div className="panel float-in">
                    <h2 className="section-title" style={{ fontSize: 24 }}>
                        Games
                    </h2>
                    <ul className="game-list">
                        {GAMES.map((g) => (
                            <li key={g.id}>
                                <Link to={`/lobby?game=${g.id}`}>
                                    <span className="game-icon" aria-hidden>
                                        {g.icon}
                                    </span>
                                    <span>
                                        <strong>{g.name}</strong>
                                        <span className="muted">{g.tagline}</span>
                                    </span>
                                    {stats?.queues?.[g.id] ? <span className="chip">{stats.queues[g.id]} waiting</span> : null}
                                </Link>
                            </li>
                        ))}
                    </ul>
                </div>
            </div>

            <div className="section-heading">
                <h2 className="section-title" style={{ fontSize: 22 }}>
                    Live server status
                </h2>
                <span className="chip">Refreshes every {REFRESH_MS / 1000}s</span>
            </div>

            {error && <p className="form-error">Could not load server statistics: {error}</p>}
            {!stats && !error && <p className="section-subtitle">Loading server statistics...</p>}

            {stats && (
                <div className="stat-grid">
                    <StatCard label="Players online" value={stats.players} />
                    <StatCard label="Live matches" value={stats.active_matches} />
                    <StatCard label="Open rooms" value={stats.rooms} />
                    <StatCard label="In queue" value={stats.queue} />
                    <StatCard label="Memory" value={stats.memory_usage} />
                    <StatCard label="CPU" value={stats.cpu_usage} />
                </div>
            )}
        </div>
    );
};

export default Homepage;
