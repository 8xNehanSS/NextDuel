import { useCallback, useEffect, useState } from "react";
import { Link, useNavigate } from "react-router";
import { useGameStore } from "../../store/gameStore";
import { useUserStore } from "../../store/userStore";
import { GameAction } from "../../utils/websocket/events";
import { socketManager } from "../../utils/websocket/socket";
import type { GameMode, PlayerState, RoundStartPayload } from "../../types/protocol";
import { RPS_ICONS, gameInfo } from "../../games";

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
    const game = gameInfo(room.mode);

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
                <span className="chip chip-muted">
                    {game.icon} {game.name}
                </span>
                <ul className="info-list" style={{ marginTop: 10 }}>
                    {game.rules.map((rule) => (
                        <li key={rule}>{rule}</li>
                    ))}
                </ul>
            </div>

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

/** The prompt, rendered to suit each game. */
const Prompt = ({ mode, text }: { mode: GameMode; text: string }) => {
    if (mode === "scramble") {
        return (
            <div className="letter-tiles" aria-label={`Scrambled letters: ${text}`}>
                {text.split(" ").map((letter, i) => (
                    <span key={i} className="letter-tile">
                        {letter}
                    </span>
                ))}
            </div>
        );
    }
    return <h2 className={`question-text${mode === "math" ? " question-math" : ""}`}>{text}</h2>;
};

const QuestionView = ({ question, mode }: { question: RoundStartPayload; mode: GameMode }) => {
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
    const isRps = mode === "rps";

    return (
        <div className="panel question-card float-in">
            <div className="game-header-chips">
                {question.category && <span className="chip">{question.category}</span>}
                {question.difficulty && <span className="chip chip-muted">{question.difficulty}</span>}
            </div>

            <div className="timer-track" aria-hidden>
                <div className={`timer-fill${urgent ? " timer-urgent" : ""}`} style={{ width: `${pct}%` }} />
            </div>

            <Prompt mode={mode} text={question.question} />

            {opponentAnswered && !locked && (
                <p className="banner banner-warn" role="status">
                    Your opponent has {isRps ? "picked" : "answered"}. You have {timer}s left!
                </p>
            )}

            <div className={`choice-grid${isRps ? " choice-grid-rps" : ""}`}>
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
                            {isRps && (
                                <span className="choice-icon" aria-hidden>
                                    {RPS_ICONS[choice]}
                                </span>
                            )}
                            {choice}
                        </button>
                    );
                })}
            </div>

            <p className="section-subtitle">
                {answer
                    ? `${isRps ? "Pick" : "Answer"} locked in. Waiting for the round to end...`
                    : pending
                      ? "Submitting..."
                      : `Click ${isRps ? "your pick" : "an answer"} or press 1-${question.choices.length}.`}
            </p>
        </div>
    );
};

const RoundResultView = ({ mode }: { mode: GameMode }) => {
    const me = useMe();
    const result = useGameStore((s) => s.lastResult)!;
    const isRps = mode === "rps";
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
            {result.summary && <p className="section-subtitle">{result.summary}</p>}
            {result.correctAnswer && (
                <p className="section-subtitle">
                    Correct answer: <strong>{result.correctAnswer}</strong>
                </p>
            )}
            <ul className="ready-list">
                {result.results.map((r) => {
                    // In RPS a tie is neither right nor wrong.
                    const tone = r.correct ? "chip-success" : isRps && !result.winnerId ? "chip-muted" : "chip-danger";
                    const label = !r.answer
                        ? "No answer"
                        : isRps
                          ? `${RPS_ICONS[r.answer] ?? ""} ${r.answer}`
                          : `${r.answer} ${r.correct ? "✓" : "✗"}`;
                    return (
                        <li key={r.playerId}>
                            <span>{r.playerId === me ? "You" : r.playerId}</span>
                            <span className={`chip ${tone}`}>{label}</span>
                        </li>
                    );
                })}
            </ul>
            {mine && !mine.answer && <p className="section-subtitle">Time ran out before you {isRps ? "picked" : "answered"}.</p>}
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
    const mode = useGameStore((s) => s.room?.mode) ?? "trivia";
    const game = gameInfo(mode);

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
        void socketManager.send(GameAction.JOIN_QUEUE, { mode });
        navigate(`/lobby?game=${mode}`);
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
            {last && over.reason === "completed" && (last.correctAnswer || last.summary) && (
                <p className="section-subtitle">
                    Last round: {last.summary} {last.correctAnswer && <strong>({last.correctAnswer})</strong>}
                </p>
            )}
            <ul className="ready-list">
                {[...over.players]
                    .sort((a, b) => b.roundsWon - a.roundsWon || b.score - a.score)
                    .map((p) => (
                        <li key={p.id}>
                            <span>{p.id === me ? `${p.id} (you)` : p.id}</span>
                            <span className="muted">
                                {p.roundsWon} rounds won
                                {mode !== "rps" && ` · ${p.score} ${game.scoredLabel.toLowerCase()}`}
                            </span>
                        </li>
                    ))}
            </ul>
            <div className="hero-actions">
                <button className="btn btn-primary" onClick={playAgain}>
                    Play {game.name} again
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
    const game = gameInfo(room.mode);

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
                    <h1 className="section-title">
                        {game.icon} {game.name}
                    </h1>
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
                            <p className="section-subtitle">Both players are ready. Setting up the duel.</p>
                        </div>
                    )}
                    {phase === "question" && question && <QuestionView key={question.round} question={question} mode={room.mode} />}
                    {phase === "round_result" && <RoundResultView mode={room.mode} />}
                    {phase === "game_over" && <GameOverView />}
                </div>
                <Scoreboard players={room.players} me={me} />
            </div>
        </div>
    );
};

export default GamePage;
