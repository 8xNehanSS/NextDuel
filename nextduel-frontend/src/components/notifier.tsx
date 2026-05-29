import { useEffect, useRef, useState } from "react";

import { appEventBus } from "../utils/eventbus";

type NotificationItem = {
    id: number;
    message: string;
    type: "info" | "error" | "success";
    description?: string;
};

const NOTIFICATION_LIMIT = 5;

const Notifier = () => {
    const [items, setItems] = useState<NotificationItem[]>([]);
    const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set());
    const nextId = useRef(1);

    useEffect(() => {
        const pushItem = (item: Omit<NotificationItem, "id">) => {
            setItems((prev) => {
                const next = [...prev, { id: nextId.current++, ...item }];
                return next.slice(-NOTIFICATION_LIMIT);
            });
        };

        const unsubscribeNotification = appEventBus.on(
            "notification",
            (payload) => {
                pushItem({ message: payload.message, type: payload.type, description: payload.description });
            }
        );

        const unsubscribeGameNotification = appEventBus.on(
            "game:notification",
            (payload) => {
                pushItem({ message: payload.message, type: "info" });
            }
        );

        return () => {
            unsubscribeNotification();
            unsubscribeGameNotification();
        };
    }, []);

    const dismiss = (id: number) => {
        setItems((prev) => prev.filter((item) => item.id !== id));
        setExpandedIds((prev) => {
            if (!prev.has(id)) {
                return prev;
            }

            const next = new Set(prev);
            next.delete(id);
            return next;
        });
    };

    const toggleExpand = (id: number) => {
        setExpandedIds((prev) => {
            const next = new Set(prev);
            if (next.has(id)) {
                next.delete(id);
            } else {
                next.add(id);
            }
            return next;
        });
    };

    const copyDescription = async (description: string | undefined) => {
        if (!description) {
            return;
        }

        try {
            await navigator.clipboard.writeText(description);
        } catch {
            // Ignore clipboard failures silently.
        }
    };

    return (
        <div className="toast-stack">
            {items.map((item) => (
                <div
                    key={item.id}
                    role="status"
                    aria-live="polite"
                    className={`toast toast-${item.type}`}
                >
                    <div style={{ display: "flex", justifyContent: "space-between", gap: 8 }}>
                        <div style={{ fontWeight: 600 }}>{item.message}</div>
                        <div className="toast-actions">
                            {item.description && (
                                <button
                                    type="button"
                                    onClick={() => toggleExpand(item.id)}
                                    className="toast-button"
                                >
                                    {expandedIds.has(item.id) ? "Hide" : "Expand"}
                                </button>
                            )}
                            <button
                                type="button"
                                onClick={() => dismiss(item.id)}
                                className="toast-button"
                            >
                                Close
                            </button>
                        </div>
                    </div>
                    {item.description && expandedIds.has(item.id) && (
                        <div style={{ marginTop: 8 }}>
                            <div style={{ fontSize: 12, color: "#64748b", lineHeight: 1.4 }}>
                                {item.description}
                            </div>
                            <button
                                type="button"
                                onClick={() => copyDescription(item.description)}
                                className="toast-button"
                                style={{ marginTop: 6 }}
                            >
                                Copy description
                            </button>
                        </div>
                    )}
                </div>
            ))}
        </div>
    );
};

export default Notifier;