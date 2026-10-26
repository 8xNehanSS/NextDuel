import { GAMES } from "../games";
import type { GameMode } from "../types/protocol";

type Props = {
    value: GameMode | undefined;
    onChange: (mode: GameMode | undefined) => void;
    allLabel?: string;
};

/** "All games" + one tab per game. undefined = all. */
const GameTabs = ({ value, onChange, allLabel = "All games" }: Props) => (
    <div className="tabs" role="tablist" aria-label="Filter by game">
        <button role="tab" aria-selected={value === undefined} className={`tab${value === undefined ? " tab-active" : ""}`} onClick={() => onChange(undefined)}>
            {allLabel}
        </button>
        {GAMES.map((g) => (
            <button
                key={g.id}
                role="tab"
                aria-selected={value === g.id}
                className={`tab${value === g.id ? " tab-active" : ""}`}
                onClick={() => onChange(g.id)}
            >
                <span aria-hidden>{g.icon}</span> {g.name}
            </button>
        ))}
    </div>
);

export default GameTabs;
