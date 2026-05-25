import { api } from "../../utils/apiclient";
import { navigationService } from "../../utils/navigationService";
import { useUserStore } from "../../store/userStore";

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

export async function login(request: LoginRequest): Promise<void> {
    const redirectParam = new URLSearchParams(window.location.search).get("redirect");
    let ep = "/login";
    if (redirectParam) {
        ep += `?redirect=${encodeURIComponent(redirectParam)}`;
    }
    const response = await api.post<{ success: boolean; redirect?: string }>(ep, request);
    const user = await getUser();
    if (user?.username) {
        useUserStore.getState().setUser(user.username);
    }
    navigationService.go(response?.redirect || redirectParam || "/");
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