import { useEffect, useState } from "react";
import { Link } from "react-router";
import { fetchServerStatistics } from "../../services/api/common";
import type { ServerStatistics } from "../../types/api";
import { errorMessage } from "../../utils/apiclient";
import { useUserStore } from "../../store/userStore";

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
                        Head-to-head trivia in real time. Answer first to put your rival on the clock.
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
                        How it works
                    </h2>
                    <ul className="info-list" style={{ marginTop: 16 }}>
                        <li>Every duel is a series of multiple-choice rounds.</li>
                        <li>The first answer in a round cuts your opponent's clock short.</li>
                        <li>Win a round by being the only one to answer correctly.</li>
                        <li>Most rounds won takes the match.</li>
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
