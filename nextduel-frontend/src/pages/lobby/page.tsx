import { socketManager } from "../../utils/websocket/socket";
import { useGameStore } from "../../store/gameStore";
import { useEffect } from "react";
import { useNavigate } from "react-router";

const LobbyPage = () => {
    const navigateTo = useNavigate();
    const gameState = useGameStore((state) => state.phase);

    useEffect(() => {
        if (gameState !== "none") {
            navigateTo("/game");
        }
    }, []);

    async function connectToServer() {
        await socketManager.send("join_queue");
    }

    return (
        <div className="flex flex-col items-center justify-center gap-3 mt-4">
            <p className="text-lg text-gray-600">Welcome to the lobby! Here you can join or create a game.</p>
            {gameState === "waiting" && <p className="text-lg text-gray-600 mt-4">Waiting for other players to join...</p>}
            {gameState === "none" && (
                <button className="bg-blue-500 hover:bg-blue-700 text-white font-bold p-2 rounded" onClick={connectToServer}>
                    Join Game
                </button>
            )}
        </div>
    );
}

export default LobbyPage;