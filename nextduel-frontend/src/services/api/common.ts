import { api, HttpError } from "../../utils/apiclient";
import type { ServerStatistics, User } from "../../types/api";

export interface LoginRequest {
    username: string;
    password: string;
}

export interface RegisterRequest {
    username: string;
    password: string;
    email: string;
}

export function fetchServerStatistics(signal?: AbortSignal): Promise<ServerStatistics> {
    return api.get<ServerStatistics>("/stats", { signal });
}

export function login(request: LoginRequest): Promise<{ username: string }> {
    return api.post<{ username: string }>("/login", request);
}

export async function logout(): Promise<void> {
    await api.post("/logout");
}

export function register(request: RegisterRequest): Promise<User> {
    return api.post<User>("/register", request);
}

/** Returns the current user, or null when not signed in. */
export async function getUser(): Promise<User | null> {
    try {
        return await api.get<User>("/user");
    } catch (error) {
        if (error instanceof HttpError && error.status === 401) {
            return null;
        }
        throw error;
    }
}
