export interface GameStart {
    question: string;
    answers: string[];
    round: number;
    players: {
        id: string;
        name: string;
    }[];
}