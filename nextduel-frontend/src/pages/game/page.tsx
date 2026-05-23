import { useState } from "react";
import { useGameStore } from "../../store/gameStore";
import { socketManager } from "../../utils/websocket/socket";

const GamePage = () => {
    const phase = useGameStore((state) => state.phase);
    const timer = useGameStore((state) => state.timer);
    const question = useGameStore((state) => state.currentQuestion);
    const round = useGameStore((state) => state.round);

    // local state
    const [answer, setAnswer] = useState("");

    async function handleReady() {
        await socketManager.send("player_ready");
    }

    async function handleAnswerSubmit() {
        if (!answer.trim()) {
            alert("Please enter an answer.");
            return;
        }
        await socketManager.send("duel_answer", {
            questionId: question?.id,
            answer: answer.trim(),
        });
    }

    return (
        <div>
            {phase === "matched" && <button onClick={handleReady}>Ready up</button>}
            <h1>Game Page</h1>
            <p>Phase: {phase}</p>
            <p>Timer: {timer}</p>
            {phase === "intermission" && <p>Intermission - get ready for the next round!</p>}
            {phase == "started" && (
                <div>
                    <p>Question: {question?.question ?? <p>No question available</p>}</p>
                    <p>Round: {round}</p>
                    <input
                        type="text"
                        placeholder="Your answer"
                        value={answer}
                        onChange={(e) => setAnswer(e.target.value)}
                    />
                    <button onClick={handleAnswerSubmit}>Submit Answer</button>
                </div>
            )}
        </div>
    );
}

export default GamePage;