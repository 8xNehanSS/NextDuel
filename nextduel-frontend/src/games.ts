// Display metadata for each game mode. IDs mirror nextduel-svr/utils/game_enums.go
// (GameMode); the server owns the rules, this file only describes them.
import type { GameMode } from "./types/protocol";

export interface GameInfo {
    id: GameMode;
    name: string;
    icon: string;
    tagline: string;
    rules: string[];
    /** Label for a scored round in results, e.g. "Correct" or "Won". */
    scoredLabel: string;
}

export const GAMES: GameInfo[] = [
    {
        id: "trivia",
        name: "Trivia Duel",
        icon: "🧠",
        tagline: "General-knowledge questions, four answers each.",
        rules: [
            "Win a round by being the only one to answer correctly.",
            "Answering first cuts your opponent's clock short.",
        ],
        scoredLabel: "Correct",
    },
    {
        id: "math",
        name: "Math Sprint",
        icon: "➗",
        tagline: "Mental arithmetic that gets harder every round.",
        rules: ["The first correct answer takes the round.", "A wrong answer locks you out of that round."],
        scoredLabel: "Correct",
    },
    {
        id: "scramble",
        name: "Word Scramble",
        icon: "🔤",
        tagline: "Unscramble the letters and spot the real word.",
        rules: ["Every option uses the same letters, but only one is a word.", "The first correct answer takes the round."],
        scoredLabel: "Correct",
    },
    {
        id: "rps",
        name: "Rock Paper Scissors",
        icon: "✊",
        tagline: "The classic, best of five. Picks stay hidden until the reveal.",
        rules: ["Rock crushes scissors, scissors cut paper, paper covers rock.", "Not picking in time loses the round."],
        scoredLabel: "Won",
    },
];

export const DEFAULT_GAME: GameMode = "trivia";

const byId = new Map(GAMES.map((g) => [g.id, g]));

export function gameInfo(id: GameMode | string | undefined): GameInfo {
    return byId.get(id as GameMode) ?? byId.get(DEFAULT_GAME)!;
}

export function isGameMode(id: string | null | undefined): id is GameMode {
    return !!id && byId.has(id as GameMode);
}

export const RPS_ICONS: Record<string, string> = { Rock: "✊", Paper: "✋", Scissors: "✌️" };
