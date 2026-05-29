import { useEffect, useState } from "react";
import {
    fetchServerStatistics,
    type ServerStatistics,
} from "../../services/api/common";

const Homepage = () => {
    const [serverStatistics, setServerStatistics] =
        useState<ServerStatistics | null>(null);

    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchStats = async () => {
            try {
                setLoading(true);
                const stats = await fetchServerStatistics();
                setServerStatistics(stats);
            } catch (error) {
                console.error("Error fetching server statistics:", error);
            } finally {
                setLoading(false);
            }
        };

        fetchStats();
    }, []);

    const StatCard = ({
        label,
        value,
    }: {
        label: string;
        value: string | number;
    }) => (
        <div className="stat-card">
            <p className="text-sm">{label}</p>
            <p className="stat-value">{value}</p>
        </div>
    );

    return (
        <div className="page-shell">
            <div className="hero">
                <div className="panel hero-card float-in">
                    <h1 className="hero-title">NextDuel control room</h1>
                    <p className="section-subtitle">
                        Keep an eye on live activity and system health before you drop into the arena.
                    </p>
                    <div className="hero-actions">
                        <span className="chip">Live telemetry</span>
                        <span className="chip chip-warn">Realtime updates</span>
                    </div>
                </div>

                <div className="panel float-in">
                    <h2 className="section-title" style={{ fontSize: 24 }}>
                        Status feed
                    </h2>
                    <p className="section-subtitle">
                        Monitor queue flow, room activity, and platform load from a single view.
                    </p>
                </div>
            </div>

            {loading && <p className="section-subtitle">Loading server statistics...</p>}

            {serverStatistics && (
                <div className="stat-grid" style={{ marginTop: 12 }}>
                    <StatCard label="Rooms" value={serverStatistics.rooms} />
                    <StatCard label="Queue" value={serverStatistics.queue} />
                    <StatCard label="Players" value={serverStatistics.players} />
                    <StatCard label="Players in Rooms" value={serverStatistics.players_in_rooms} />
                    <StatCard label="Players in Queue" value={serverStatistics.players_in_queue} />
                    <StatCard label="Memory Usage" value={serverStatistics.memory_usage} />
                    <StatCard label="CPU Usage" value={serverStatistics.cpu_usage} />
                </div>
            )}
        </div>
    );
};

export default Homepage;