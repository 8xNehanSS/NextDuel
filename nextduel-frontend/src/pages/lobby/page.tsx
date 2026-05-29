import { GameAction } from "../../utils/websocket/events";
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
        await socketManager.send(GameAction.JOIN_QUEUE);
    }

    return (
        <div className="page-shell">
            <div className="hero">
                <div className="panel hero-card float-in">
                    <h1 className="hero-title">Queue up for a live duel.</h1>
                    <p className="section-subtitle">
                        Match with a rival, solve the prompt, and climb the leaderboard in real time.
                    </p>

                    <div className="hero-actions">
                        {gameState === "none" && (
                            <button className="btn btn-primary" onClick={connectToServer}>
                                Join matchmaking
                            </button>
                        )}
                        {gameState === "waiting" && (
                            <span className="chip chip-warn">Searching for an opponent...</span>
                        )}
                    </div>
                </div>

                <div className="panel float-in">
                    <span className="chip">Realtime PvP</span>
                    <h2 className="section-title" style={{ fontSize: 24, marginTop: 12 }}>
                        What happens next
                    </h2>
                    <p className="section-subtitle">
                        Once matched, you will head straight into the duel room with a countdown timer.
                    </p>
                    <ul className="info-list" style={{ marginTop: 16 }}>
                        <li>Answer quickly to trigger the fast-duel timer.</li>
                        <li>Win rounds to secure the match.</li>
                        <li>Track your progress in the live scoreboard.</li>
                    </ul>
                </div>
            </div>
        </div>
    );
}

export default LobbyPage;