import type { GameAction } from "./events";
import { handlers } from "./handler";

const random4DigitId = () => Math.floor(1000 + Math.random() * 9000).toString();

type ServerMessage<T = any> = {
    type: GameAction;
    payload?: T;
    data?: T;
};

class SocketManager {
    private socket: WebSocket | null = null;
    private connecting: Promise<void> | null = null;

    connect(): Promise<void> {
        if (this.socket?.readyState === WebSocket.OPEN) {
            return Promise.resolve();
        }

        if (this.connecting) return this.connecting;

        this.connecting = new Promise((resolve) => {
            this.socket = new WebSocket(`/ws?token=${random4DigitId()}`);
            this.socket.onopen = () => {
                console.log("Connected");
                resolve();
                this.connecting = null;
            };

            this.socket.onmessage = (event) => {
                const message: ServerMessage = JSON.parse(event.data);
                this.handleMessage(message);
            };

            this.socket.onclose = () => {
                console.log("Disconnected");
                this.socket = null;
                this.connecting = null;
            };

            this.socket.onerror = (err) => {
                console.error("WS error", err);
            };
        });

        return this.connecting;
    }

    async send(type: GameAction, payload?: any) {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            await this.connect();
        }

        this.socket?.send(
            JSON.stringify({
                type,
                data: payload,
            })
        );
    }

    private handleMessage(message: ServerMessage) {
        const handler = handlers[message.type];

        if (!handler) {
            console.warn("No handler for:", message.type);
            return;
        }

        handler(message.payload ?? message.data);
    }
}

export const socketManager = new SocketManager();