import { api } from "../../utils/apiclient";

export interface ServerStatistics {
    rooms: number;
    queue: number;
    players: number;
    players_in_rooms: number;
    players_in_queue: number;
    memory_usage: string;
    cpu_usage: string;
}

export async function fetchServerStatistics(): Promise<ServerStatistics> {
    const response = await api.get("/health");
    return response as ServerStatistics;
}

export interface LoginRequest {
    username: string;
    password: string;
}

export async function login(request: LoginRequest) {
    await api.post("/login", request);
}