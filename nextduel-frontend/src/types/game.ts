export interface GameStart {
    quest: string;
    round: number;
    players: {
        id: string;
        name: string;
    }[];
}