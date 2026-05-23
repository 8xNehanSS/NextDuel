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
        <div className="border rounded-lg p-4 shadow-sm">
            <p className="text-sm">{label}</p>
            <p className="text-lg font-semibold mt-1">{value}</p>
        </div>
    );

    return (
        <div className="px-6 py-10">
            <div className="max-w-5xl mx-auto space-y-8">

                <div className="space-y-1">
                    <p className="text-sm">Welcome to the home page!</p>
                </div>

                {loading && (
                    <p className="text-sm">Loading server statistics...</p>
                )}

                {serverStatistics && (
                    <div className="space-y-4">

                        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                            <StatCard
                                label="Rooms"
                                value={serverStatistics.rooms}
                            />
                            <StatCard
                                label="Queue"
                                value={serverStatistics.queue}
                            />
                            <StatCard
                                label="Players"
                                value={serverStatistics.players}
                            />
                            <StatCard
                                label="Players in Rooms"
                                value={serverStatistics.players_in_rooms}
                            />
                            <StatCard
                                label="Players in Queue"
                                value={serverStatistics.players_in_queue}
                            />
                            <StatCard
                                label="Memory Usage"
                                value={serverStatistics.memory_usage}
                            />
                            <StatCard
                                label="CPU Usage"
                                value={serverStatistics.cpu_usage}
                            />
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default Homepage;