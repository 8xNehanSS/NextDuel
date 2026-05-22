import { useEffect, useState } from "react";
import { fetchServerStatistics, type ServerStatistics } from "../../services/api/common";

const homepage = () => {
    const [serverStatistics, setServerStatistics] = useState<ServerStatistics | null>(null);

    useEffect(() => {
        const fetchStats = async () => {
            try {
                const stats = await fetchServerStatistics();
                setServerStatistics(stats);
            } catch (error) {
                console.error("Error fetching server statistics:", error);
            }
        };
        fetchStats();
    }, []);

    return (
        <div>
            <h1>Home Page</h1>
            <p>Welcome to the home page!</p>
            {serverStatistics && (
                <div>
                    <h2>Server Statistics</h2>
                    <p>Rooms: {serverStatistics.rooms}</p>
                    <p>Queue: {serverStatistics.queue}</p>
                    <p>Players: {serverStatistics.players}</p>
                    <p>Players in Rooms: {serverStatistics.players_in_rooms}</p>
                    <p>Players in Queue: {serverStatistics.players_in_queue}</p>
                    <p>Memory Usage: {serverStatistics.memory_usage}</p>
                    <p>CPU Usage: {serverStatistics.cpu_usage}</p>
                </div>
            )}
        </div>
    );
}

export default homepage;