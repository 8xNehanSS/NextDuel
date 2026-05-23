import { api } from "../../utils/apiclient";
import { navigationService } from "../../utils/navigationService";

export interface ServerResponse<T> {
    success: boolean;
    message?: string;
    data?: T;
}

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

export interface UserRequest {
    username: string;
}

export async function getUser(): Promise<UserRequest> {
    const response = await api.get("/user") as ServerResponse<UserRequest>;
    return response.data as UserRequest;
}

export async function logout() {
    await api.post("/logout");
}

export interface RegisterRequest {
    username: string;
    password: string;
    email: string;
}

export async function register(request: RegisterRequest) {
    const response = await api.post("/register", request) as ServerResponse<any>;
    if (!response.success) {
        throw new Error(response.message);
    } else {
        navigationService.go("/login");
    }
}