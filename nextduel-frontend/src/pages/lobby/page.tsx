import { useEffect, useState, type FormEvent } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { GameAction } from "../../utils/websocket/events";
import { socketManager } from "../../utils/websocket/socket";
import { useGameStore } from "../../store/gameStore";
import { DEFAULT_GAME, GAMES, gameInfo, isGameMode } from "../../games";
import type { GameMode } from "../../types/protocol";

const CODE_LENGTH = 6;
const LAST_GAME_KEY = "nextduel:lastGame";

// Remembering the last game is a convenience; storage may be unavailable.
function initialGame(param: string | null): GameMode {
    if (isGameMode(param)) return param;
    try {
        const stored = window.localStorage.getItem(LAST_GAME_KEY);
        if (isGameMode(stored)) return stored;
    } catch {
        // ignore
    }
    return DEFAULT_GAME;
}

const LobbyPage = () => {
    const navigate = useNavigate();
    const [params] = useSearchParams();
    const phase = useGameStore((s) => s.phase);
    const queueMode = useGameStore((s) => s.queueMode);
    const queuePosition = useGameStore((s) => s.queuePosition);
    const reset = useGameStore((s) => s.reset);
    const [selected, setSelected] = useState<GameMode>(() => initialGame(params.get("game")));
    const [code, setCode] = useState("");

    useEffect(() => {
        if (phase === "game_over") {
            reset();
        } else if (phase !== "idle" && phase !== "queued") {
            navigate("/game", { replace: true });
        }
    }, [phase, navigate, reset]);

    function choose(mode: GameMode) {
        setSelected(mode);
        try {
            window.localStorage.setItem(LAST_GAME_KEY, mode);
        } catch {
            // ignore
        }
    }

    const queued = phase === "queued";
    const game = gameInfo(selected);
    const queuedGame = gameInfo(queueMode ?? selected);

    function joinRoom(e: FormEvent<HTMLFormElement>) {
        e.preventDefault();
        const roomId = code.trim().toUpperCase();
        if (roomId.length !== CODE_LENGTH) return;
        void socketManager.send(GameAction.JOIN_ROOM, { roomId });
    }

    return (
        <div className="page-shell">
            <h1 className="section-title">Choose a game</h1>
            <p className="section-subtitle">Every game is a head-to-head duel. Most rounds won takes the match.</p>

            <div className="game-picker" role="radiogroup" aria-label="Game">
                {GAMES.map((g) => (
                    <button
                        key={g.id}
                        role="radio"
                        aria-checked={selected === g.id}
                        className={`game-card${selected === g.id ? " game-card-selected" : ""}`}
                        disabled={queued}
                        onClick={() => choose(g.id)}
                    >
                        <span className="game-icon" aria-hidden>
                            {g.icon}
                        </span>
                        <strong>{g.name}</strong>
                        <span className="muted">{g.tagline}</span>
                    </button>
                ))}
            </div>

            <div className="hero">
                <div className="panel hero-card float-in">
                    <span className="chip">Ranked matchmaking</span>
                    <h2 className="hero-title" style={{ marginTop: 12, fontSize: 34 }}>
                        {game.icon} {game.name}
                    </h2>
                    <ul className="info-list" style={{ marginTop: 12 }}>
                        {game.rules.map((rule) => (
                            <li key={rule}>{rule}</li>
                        ))}
                    </ul>

                    <div className="hero-actions">
                        {queued ? (
                            <>
                                <span className="chip chip-warn pulse">
                                    Searching for a {queuedGame.name} opponent
                                    {queuePosition > 1 ? ` · #${queuePosition} in queue` : "..."}
                                </span>
                                <button className="btn" onClick={() => void socketManager.send(GameAction.LEAVE_QUEUE)}>
                                    Cancel
                                </button>
                            </>
                        ) : (
                            <button className="btn btn-primary" onClick={() => void socketManager.send(GameAction.JOIN_QUEUE, { mode: selected })}>
                                Find a {game.name} match
                            </button>
                        )}
                    </div>
                </div>

                <div className="panel float-in">
                    <span className="chip">Private room</span>
                    <h2 className="section-title" style={{ fontSize: 24, marginTop: 12 }}>
                        Challenge a friend
                    </h2>
                    <p className="section-subtitle">
                        Create a {game.name} room and share its code, or enter a code you were sent. The room's creator picks the game.
                    </p>

                    <div className="hero-actions">
                        <button
                            className="btn btn-primary"
                            disabled={queued}
                            onClick={() => void socketManager.send(GameAction.CREATE_ROOM, { mode: selected })}
                        >
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
