import { useCallback, useEffect, useState } from "react";
import { Link, useNavigate } from "react-router";
import { useGameStore } from "../../store/gameStore";
import { useUserStore } from "../../store/userStore";
import { GameAction } from "../../utils/websocket/events";
import { socketManager } from "../../utils/websocket/socket";
import type { PlayerState, RoundStartPayload } from "../../types/protocol";

const choiceKeys = ["1", "2", "3", "4"];

function useMe() {
    return useUserStore((s) => s.username) ?? "";
}

// ---------------------------------------------------------------------------

const Scoreboard = ({ players, me }: { players: PlayerState[]; me: string }) => (
    <div className="panel float-in">
        <span className="chip">Scoreboard</span>
        <div className="score-list" style={{ marginTop: 16 }}>
            {players.map((p) => (
                <div key={p.id} className={`score-row${p.id === me ? " score-row-me" : ""}`}>
                    <span>
                        {p.id}
                        {p.id === me && <span className="muted"> (you)</span>}
                    </span>
                    <span className="score-values">
                        <strong>{p.score}</strong>
                        <span className="muted"> pts · {p.roundsWon} won</span>
                    </span>
                </div>
            ))}
            {players.length < 2 && <div className="score-row muted">Waiting for an opponent...</div>}
        </div>
    </div>
);

const LobbyView = () => {
    const me = useMe();
    const room = useGameStore((s) => s.room)!;
    const [copied, setCopied] = useState(false);
    const self = room.players.find((p) => p.id === me);
    const full = room.players.length === 2;

    async function copyCode() {
        try {
            await navigator.clipboard.writeText(room.roomId);
            setCopied(true);
            window.setTimeout(() => setCopied(false), 1500);
        } catch {
            // Clipboard can be unavailable on insecure origins; the code is visible anyway.
        }
    }

    return (
        <div className="panel float-in stack">
            {room.private && (
                <div>
                    <p className="form-label">Room code</p>
                    <div className="room-code-row">
                        <span className="room-code">{room.roomId}</span>
                        <button className="btn btn-ghost" onClick={copyCode}>
                            {copied ? "Copied!" : "Copy"}
                        </button>
                    </div>
                    {!full && <p className="section-subtitle">Share this code with a friend so they can join.</p>}
                </div>
            )}

            <div>
                <h2 className="section-title" style={{ fontSize: 24 }}>
                    {full ? "Opponent found" : "Waiting for an opponent"}
                </h2>
                <ul className="ready-list">
                    {room.players.map((p) => (
                        <li key={p.id}>
                            <span>{p.id === me ? `${p.id} (you)` : p.id}</span>
                            <span className={`chip ${p.ready ? "" : "chip-warn"}`}>{p.ready ? "Ready" : "Not ready"}</span>
                        </li>
                    ))}
                </ul>
            </div>

            <div className="hero-actions">
                <button
                    className="btn btn-primary"
                    disabled={!full || self?.ready}
                    onClick={() => void socketManager.send(GameAction.PLAYER_READY)}
                >
                    {self?.ready ? "Waiting for opponent..." : "Ready up"}
                </button>
            </div>
        </div>
    );
};

const QuestionView = ({ question }: { question: RoundStartPayload }) => {
    const timer = useGameStore((s) => s.timer);
    const answer = useGameStore((s) => s.answer);
    const pending = useGameStore((s) => s.pendingAnswer);
    const opponentAnswered = useGameStore((s) => s.opponentAnswered);
    const answerPending = useGameStore((s) => s.answerPending);
    const locked = answer !== null || pending !== null;

    const choose = useCallback(
        (choice: string) => {
            if (locked) return;
            answerPending(choice);
            void socketManager.send(GameAction.DUEL_ANSWER, { round: question.round, answer: choice });
        },
        [locked, question.round, answerPending],
    );

    useEffect(() => {
        function onKey(e: KeyboardEvent) {
            const index = choiceKeys.indexOf(e.key);
            if (index >= 0 && index < question.choices.length) {
                choose(question.choices[index]);
            }
        }
        window.addEventListener("keydown", onKey);
        return () => window.removeEventListener("keydown", onKey);
    }, [choose, question.choices]);

    const pct = Math.max(0, Math.min(100, (timer / question.duration) * 100));
    const urgent = timer <= 5;

    return (
        <div className="panel question-card float-in">
            <div className="game-header-chips">
                {question.category && <span className="chip">{question.category}</span>}
                {question.difficulty && <span className="chip chip-muted">{question.difficulty}</span>}
            </div>

            <div className="timer-track" aria-hidden>
                <div className={`timer-fill${urgent ? " timer-urgent" : ""}`} style={{ width: `${pct}%` }} />
            </div>

            <h2 className="question-text">{question.question}</h2>

            {opponentAnswered && !locked && (
                <p className="banner banner-warn" role="status">
                    Your opponent has answered. You have {timer}s left!
                </p>
            )}

            <div className="choice-grid">
                {question.choices.map((choice, i) => {
                    const selected = choice === answer || choice === pending;
                    return (
                        <button
                            key={choice}
                            className={`choice${selected ? " choice-selected" : ""}`}
                            disabled={locked && !selected}
                            aria-pressed={selected}
                            onClick={() => choose(choice)}
                        >
                            <span className="choice-key">{choiceKeys[i]}</span>
                            {choice}
                        </button>
                    );
                })}
            </div>

            <p className="section-subtitle">
                {answer
                    ? "Answer locked in. Waiting for the round to end..."
                    : pending
                      ? "Submitting..."
                      : "Click an answer or press 1-4."}
            </p>
        </div>
    );
};

const RoundResultView = () => {
    const me = useMe();
    const result = useGameStore((s) => s.lastResult)!;
    const nextRound = useGameStore((s) => s.nextRound);
    const timer = useGameStore((s) => s.timer);
    const mine = result.results.find((r) => r.playerId === me);

    let headline = "Nobody takes this round";
    if (result.winnerId === me) headline = "You win the round!";
    else if (result.winnerId) headline = `${result.winnerId} wins the round`;

    return (
        <div className="panel float-in stack">
            <span className={`chip ${result.winnerId === me ? "" : "chip-warn"}`}>Round {result.round} result</span>
            <h2 className="section-title" style={{ fontSize: 28 }}>
                {headline}
            </h2>
            <p className="section-subtitle">
                Correct answer: <strong>{result.correctAnswer}</strong>
            </p>
            <ul className="ready-list">
                {result.results.map((r) => (
                    <li key={r.playerId}>
                        <span>{r.playerId === me ? "You" : r.playerId}</span>
                        <span className={`chip ${r.correct ? "chip-success" : "chip-danger"}`}>
                            {r.answer ? `${r.answer} ${r.correct ? "✓" : "✗"}` : "No answer"}
                        </span>
                    </li>
                ))}
            </ul>
            {mine && !mine.answer && <p className="section-subtitle">Time ran out before you answered.</p>}
            {nextRound && <p className="section-subtitle">Round {nextRound} starts in {timer}s...</p>}
        </div>
    );
};

const GameOverView = () => {
    const me = useMe();
    const navigate = useNavigate();
    const reset = useGameStore((s) => s.reset);
    const over = useGameStore((s) => s.gameOver)!;
    const last = useGameStore((s) => s.lastResult);

    let title = "It's a draw";
    let tone = "draw";
    if (over.winnerId === me) {
        title = "Victory!";
        tone = "win";
    } else if (over.winnerId) {
        title = "Defeat";
        tone = "loss";
    }

    function playAgain() {
        reset();
        void socketManager.send(GameAction.JOIN_QUEUE);
        navigate("/lobby");
    }

    function backToLobby() {
        reset();
        navigate("/lobby");
    }

    return (
        <div className={`panel float-in stack result-${tone}`}>
            <h2 className="hero-title">{title}</h2>
            {over.reason === "forfeit" && (
                <p className="section-subtitle">
                    {over.winnerId === me ? "Your opponent left the match." : "The match ended early."}
                </p>
            )}
            {last && over.reason === "completed" && (
                <p className="section-subtitle">
                    Final question answer: <strong>{last.correctAnswer}</strong>
                </p>
            )}
            <ul className="ready-list">
                {[...over.players]
                    .sort((a, b) => b.roundsWon - a.roundsWon || b.score - a.score)
                    .map((p) => (
                        <li key={p.id}>
                            <span>{p.id === me ? `${p.id} (you)` : p.id}</span>
                            <span className="muted">
                                {p.roundsWon} rounds · {p.score} correct
                            </span>
                        </li>
                    ))}
            </ul>
            <div className="hero-actions">
                <button className="btn btn-primary" onClick={playAgain}>
                    Play again
                </button>
                <button className="btn" onClick={backToLobby}>
                    Back to lobby
                </button>
                <Link className="btn btn-ghost" to="/history" onClick={reset}>
                    View history
                </Link>
            </div>
        </div>
    );
};

// ---------------------------------------------------------------------------

const GamePage = () => {
    const me = useMe();
    const navigate = useNavigate();
    const phase = useGameStore((s) => s.phase);
    const room = useGameStore((s) => s.room);
    const question = useGameStore((s) => s.question);
    const maxRounds = useGameStore((s) => s.maxRounds);
    const timer = useGameStore((s) => s.timer);
    const reset = useGameStore((s) => s.reset);

    if (phase === "idle" || phase === "queued" || !room) {
        return (
            <div className="page-shell">
                <div className="panel empty-state">
                    <h1 className="section-title">No active duel</h1>
                    <p className="section-subtitle">Find a match or create a private room to start playing.</p>
                    <div className="hero-actions" style={{ justifyContent: "center" }}>
                        <Link to="/lobby" className="btn btn-primary">
                            Go to lobby
                        </Link>
                    </div>
                </div>
            </div>
        );
    }

    const inMatch = phase === "starting" || phase === "question" || phase === "round_result";

    function leave() {
        if (inMatch && !window.confirm("Leaving now forfeits the match. Leave anyway?")) {
            return;
        }
        void socketManager.send(GameAction.LEAVE_ROOM);
        reset();
        navigate("/lobby");
    }

    return (
        <div className="page-shell">
            <div className="game-header">
                <div>
                    <h1 className="section-title">Duel room</h1>
                    <p className="section-subtitle">
                        {room.players.map((p) => (p.id === me ? "You" : p.id)).join(" vs ")}
                    </p>
                </div>
                <div className="game-header-chips">
                    {room.private && <span className="chip chip-muted">Code {room.roomId}</span>}
                    {(phase === "question" || phase === "round_result") && question && (
                        <span className="chip">
                            Round {question.round} / {maxRounds}
                        </span>
                    )}
                    {phase === "question" && <span className={`chip ${timer <= 5 ? "chip-danger" : "chip-warn"}`}>{timer}s</span>}
                    {phase !== "game_over" && (
                        <button className="btn btn-ghost" onClick={leave}>
                            Leave
                        </button>
                    )}
                </div>
            </div>

            <div className="game-grid">
                <div>
                    {phase === "lobby" && <LobbyView />}
                    {phase === "starting" && (
                        <div className="panel float-in empty-state">
                            <h2 className="section-title pulse">Get ready...</h2>
                            <p className="section-subtitle">Both players are ready. Loading questions.</p>
                        </div>
                    )}
                    {phase === "question" && question && <QuestionView key={question.round} question={question} />}
                    {phase === "round_result" && <RoundResultView />}
                    {phase === "game_over" && <GameOverView />}
                </div>
                <Scoreboard players={room.players} me={me} />
            </div>
        </div>
    );
};

export default GamePage;
