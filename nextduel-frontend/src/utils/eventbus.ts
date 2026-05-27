type EventHandler<TPayload> = (payload: TPayload) => void;

type EventMap = Record<string, unknown>;

class EventBus<TEvents extends EventMap> {
	private listeners: Map<keyof TEvents, Set<EventHandler<TEvents[keyof TEvents]>>>;

	constructor() {
		this.listeners = new Map();
	}

	on<TKey extends keyof TEvents>(
		event: TKey,
		handler: EventHandler<TEvents[TKey]>
	): () => void {
		const handlers = this.getOrCreateHandlers(event);
		handlers.add(handler as EventHandler<TEvents[keyof TEvents]>);

		return () => this.off(event, handler);
	}

	listen<TKey extends keyof TEvents>(
		event: TKey,
		handler: EventHandler<TEvents[TKey]>
	): () => void {
		return this.on(event, handler);
	}

	off<TKey extends keyof TEvents>(
		event: TKey,
		handler: EventHandler<TEvents[TKey]>
	): void {
		const handlers = this.listeners.get(event);
		if (!handlers) {
			return;
		}

		handlers.delete(handler as EventHandler<TEvents[keyof TEvents]>);
		if (handlers.size === 0) {
			this.listeners.delete(event);
		}
	}

	fire<TKey extends keyof TEvents>(event: TKey, payload: TEvents[TKey]): void {
		const handlers = this.listeners.get(event);
		if (!handlers) {
			return;
		}

		for (const handler of handlers) {
			(handler as EventHandler<TEvents[TKey]>)(payload);
		}
	}

	emit<TKey extends keyof TEvents>(event: TKey, payload: TEvents[TKey]): void {
		this.fire(event, payload);
	}

	clear(event?: keyof TEvents): void {
		if (event) {
			this.listeners.delete(event);
			return;
		}

		this.listeners.clear();
	}

	private getOrCreateHandlers<TKey extends keyof TEvents>(
		event: TKey
	): Set<EventHandler<TEvents[keyof TEvents]>> {
		const existing = this.listeners.get(event);
		if (existing) {
			return existing;
		}

		const handlers = new Set<EventHandler<TEvents[keyof TEvents]>>();
		this.listeners.set(event, handlers);
		return handlers;
	}
}

type AppEventMap = {
	// Fill in your app events and payloads here.
	// Example: "user:login": { userId: string };
    "notification": { message: string; type: "info" | "error" | "success"; description?: string };
    "game:notification": { message: string };
};

const appEventBus = new EventBus<AppEventMap>();

export type { EventMap, EventHandler, AppEventMap };
export { EventBus, appEventBus };
