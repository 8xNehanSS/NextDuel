import { useState } from "react";
import { useGameStore } from "../../store/gameStore";
import { GameAction } from "../../utils/websocket/events";
import { socketManager } from "../../utils/websocket/socket";

const GamePage = () => {
    const phase = useGameStore((state) => state.phase);
    const timer = useGameStore((state) => state.timer);
    const question = useGameStore((state) => state.currentQuestion);
    const round = useGameStore((state) => state.round);
    const me = useGameStore((state) => state.me);
    const opponent = useGameStore((state) => state.opponent);
    const ready = useGameStore((state) => state.ready);
    const answered = useGameStore((state) => state.answered);

    // local state
    const [answer, setAnswer] = useState("");

    async function handleReady() {
        await socketManager.send(GameAction.PLAYER_READY);
    }

    async function handleAnswerSubmit() {
        if (!answer.trim()) {
            alert("Please enter an answer.");
            return;
        }
        await socketManager.send(GameAction.DUEL_ANSWER, {
            questionId: question?.id,
            answer: answer.trim(),
        });
        setAnswer(""); // Clear the input field after submission
    }

    return (
        <div className="page-shell">
            <div className="game-header">
                <div>
                    <h1 className="section-title">Duel Room</h1>
                    {phase !== "none" && <p className="section-subtitle">Phase: {phase}</p>}
                    {phase === "none" && <p className="section-subtitle">No active game found</p>}
                </div>
                <div className="game-header-chips">
                    <span className="chip">Round {round}</span>
                    <span className="chip chip-warn">Timer {timer}s</span>
                </div>
            </div>

            {phase === "matched" && (
                <div className="panel float-in">
                    <h2 className="section-title" style={{ fontSize: 24 }}>
                        Opponent found
                    </h2>
                    <p className="section-subtitle">Confirm when you are ready to start the duel.</p>
                    <div className="hero-actions">
                        <button className="btn btn-primary" onClick={handleReady}>
                            {ready ? "Waiting for opponent..." : "Ready up!"}
                        </button>
                    </div>
                </div>
            )}

            {phase === "intermission" && (
                <div className="panel float-in" style={{ marginTop: 20 }}>
                    <span className="chip chip-warn">Intermission</span>
                    <p className="section-subtitle" style={{ marginTop: 10 }}>
                        Reset, review the scoreboard, and prepare for the next round.
                    </p>
                </div>
            )}

            {phase === "started" && (
                <div className="game-grid" style={{ marginTop: 20 }}>
                    <div className="panel question-card float-in">
                        <span className="chip">Live Question</span>
                        <h2 className="section-title" style={{ fontSize: 28 }}>
                            {question?.question ?? "No question available"}
                        </h2>
                        {answered ? (
                            <p className="section-subtitle" style={{ marginTop: 10, color: "green" }}>
                                Answer submitted! Waiting for results...
                            </p>
                        ) : (
                            <>
                                <div>
                                    {question?.answers.map((ans, index) => (
                                        <div key={index} className="answer-option">
                                            <span>{String.fromCharCode(65 + index)}. {ans}</span>
                                        </div>)
                                    )}
                                </div>
                                <div>
                                    <input
                                        className="input-field"
                                        type="text"
                                        placeholder="Type your answer"
                                        value={answer}
                                        onChange={(e) => setAnswer(e.target.value)}
                                    />
                                </div>
                                <div className="hero-actions">
                                    <button className="btn btn-primary" onClick={handleAnswerSubmit}>
                                        Submit answer
                                    </button>
                                </div>
                            </>
                        )}
                    </div>

                    <div className="panel float-in">
                        <span className="chip">Scoreboard</span>
                        <div className="score-list" style={{ marginTop: 16 }}>
                            <div className="score-row">
                                <span>You</span>
                                <strong>{me?.score ?? 0}</strong>
                            </div>
                            <div className="score-row">
                                <span>Opponent</span>
                                <strong>{opponent?.score ?? 0}</strong>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}

export default GamePage;