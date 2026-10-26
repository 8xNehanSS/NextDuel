import { GameAction, type ClientEventType, type ClientEvents, type ServerEventType, type ServerEvents } from "./events";
import { handlers } from "./handler";
import { IN_GAME_PHASES, useGameStore } from "../../store/gameStore";
import { appEventBus } from "../eventbus";
import { navigationService } from "../navigationService";

type ServerMessage = { [K in ServerEventType]: { type: K; data: ServerEvents[K] } }[ServerEventType];

// Keeps idle connections alive through proxies with short idle timeouts.
const HEARTBEAT_MS = 25_000;

function socketUrl(): string {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    return `${protocol}//${window.location.host}/api/ws`;
}

class SocketManager {
    private socket: WebSocket | null = null;
    private connecting: Promise<void> | null = null;
    private heartbeat: number | undefined;
    private closingIntentionally = false;

    connect(): Promise<void> {
        if (this.socket?.readyState === WebSocket.OPEN) {
            return Promise.resolve();
        }
        if (this.connecting) {
            return this.connecting;
        }

        this.closingIntentionally = false;
        this.connecting = new Promise<void>((resolve, reject) => {
            const socket = new WebSocket(socketUrl());
            this.socket = socket;
            let opened = false;

            socket.onopen = () => {
                opened = true;
                this.connecting = null;
                useGameStore.getState().setConnected(true);
                this.startHeartbeat();
                resolve();
            };

            socket.onmessage = (event) => this.handleMessage(event.data);

            socket.onclose = () => {
                this.stopHeartbeat();
                this.socket = null;
                this.connecting = null;
                useGameStore.getState().setConnected(false);
                if (!opened) {
                    reject(new Error("Could not connect to the game server."));
                    return;
                }
                this.handleUnexpectedClose();
            };
        });

        return this.connecting;
    }

    async send<K extends ClientEventType>(type: K, ...payload: ClientEvents[K] extends undefined ? [] : [ClientEvents[K]]) {
        try {
            await this.connect();
        } catch (error) {
            appEventBus.fire("notification", { message: (error as Error).message, type: "error" });
            throw error;
        }
        this.socket?.send(JSON.stringify({ type, data: payload[0] }));
    }

    /** Close deliberately (e.g. on logout) without the "connection lost" flow. */
    disconnect() {
        this.closingIntentionally = true;
        this.socket?.close();
    }

    private handleUnexpectedClose() {
        if (this.closingIntentionally) {
            useGameStore.getState().reset();
            return;
        }
        if (IN_GAME_PHASES.has(useGameStore.getState().phase)) {
            appEventBus.fire("notification", {
                message: "Connection lost",
                description: "You were disconnected from the game server, so the current game has ended.",
                type: "error",
            });
            useGameStore.getState().reset();
            navigationService.go("/lobby");
        }
    }

    private handleMessage(raw: unknown) {
        let message: ServerMessage;
        try {
            message = JSON.parse(String(raw)) as ServerMessage;
        } catch {
            console.warn("Ignoring malformed websocket message", raw);
            return;
        }
        const handler = handlers[message.type] as ((data: unknown) => void) | undefined;
        if (handler) {
            handler(message.data);
        } else if (message.type !== GameAction.PONG) {
            console.warn("No handler for websocket message:", message.type);
        }
    }

    private startHeartbeat() {
        this.stopHeartbeat();
        this.heartbeat = window.setInterval(() => {
            if (this.socket?.readyState === WebSocket.OPEN) {
                this.socket.send(JSON.stringify({ type: GameAction.PING }));
            }
        }, HEARTBEAT_MS);
    }

    private stopHeartbeat() {
        window.clearInterval(this.heartbeat);
        this.heartbeat = undefined;
    }
}

export const socketManager = new SocketManager();
