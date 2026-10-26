import { useEffect, useState, type FormEvent } from "react";
import { useNavigate } from "react-router";
import { GameAction } from "../../utils/websocket/events";
import { socketManager } from "../../utils/websocket/socket";
import { useGameStore } from "../../store/gameStore";

const CODE_LENGTH = 6;

const LobbyPage = () => {
    const navigate = useNavigate();
    const phase = useGameStore((s) => s.phase);
    const queuePosition = useGameStore((s) => s.queuePosition);
    const reset = useGameStore((s) => s.reset);
    const [code, setCode] = useState("");

    useEffect(() => {
        if (phase === "game_over") {
            reset();
        } else if (phase !== "idle" && phase !== "queued") {
            navigate("/game", { replace: true });
        }
    }, [phase, navigate, reset]);

    const queued = phase === "queued";

    function joinRoom(e: FormEvent<HTMLFormElement>) {
        e.preventDefault();
        const roomId = code.trim().toUpperCase();
        if (roomId.length !== CODE_LENGTH) return;
        void socketManager.send(GameAction.JOIN_ROOM, { roomId });
    }

    return (
        <div className="page-shell">
            <div className="hero">
                <div className="panel hero-card float-in">
                    <span className="chip">Ranked matchmaking</span>
                    <h1 className="hero-title" style={{ marginTop: 12 }}>
                        Queue up for a live duel.
                    </h1>
                    <p className="section-subtitle">
                        Get paired with the next available rival. Wins count toward the leaderboard.
                    </p>

                    <div className="hero-actions">
                        {queued ? (
                            <>
                                <span className="chip chip-warn pulse">
                                    Searching for an opponent{queuePosition > 1 ? ` · #${queuePosition} in queue` : "..."}
                                </span>
                                <button className="btn" onClick={() => void socketManager.send(GameAction.LEAVE_QUEUE)}>
                                    Cancel
                                </button>
                            </>
                        ) : (
                            <button className="btn btn-primary" onClick={() => void socketManager.send(GameAction.JOIN_QUEUE)}>
                                Find a match
                            </button>
                        )}
                    </div>
                </div>

                <div className="panel float-in">
                    <span className="chip">Private room</span>
                    <h2 className="section-title" style={{ fontSize: 24, marginTop: 12 }}>
                        Challenge a friend
                    </h2>
                    <p className="section-subtitle">Create a room and share its code, or enter a code you were sent.</p>

                    <div className="hero-actions">
                        <button className="btn btn-primary" disabled={queued} onClick={() => void socketManager.send(GameAction.CREATE_ROOM)}>
                            Create room
                        </button>
                    </div>

                    <form className="join-form" onSubmit={joinRoom}>
                        <input
                            className="input-field room-code-input"
                            placeholder="ROOM CODE"
                            aria-label="Room code"
                            maxLength={CODE_LENGTH}
                            value={code}
                            disabled={queued}
                            onChange={(e) => setCode(e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, ""))}
                        />
                        <button type="submit" className="btn" disabled={queued || code.length !== CODE_LENGTH}>
                            Join
                        </button>
                    </form>
                </div>
            </div>
        </div>
    );
};

export default LobbyPage;
