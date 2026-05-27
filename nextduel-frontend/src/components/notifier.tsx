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
        <div
            style={{
                position: "fixed",
                right: 16,
                bottom: 16,
                display: "flex",
                flexDirection: "column",
                gap: 8,
                alignItems: "flex-end",
                zIndex: 1000,
            }}
        >
            {items.map((item) => (
                <div
                    key={item.id}
                    role="status"
                    aria-live="polite"
                    style={{
                        minWidth: 220,
                        maxWidth: 320,
                        padding: "10px 12px",
                        borderRadius: 10,
                        color: "#0f172a",
                        background:
                            item.type === "success"
                                ? "#bbf7d0"
                                : item.type === "error"
                                ? "#fecaca"
                                : "#e0f2fe",
                        boxShadow: "0 6px 18px rgba(15, 23, 42, 0.15)",
                        fontSize: 14,
                    }}
                >
                    <div style={{ display: "flex", justifyContent: "space-between", gap: 8 }}>
                        <div style={{ fontWeight: 600 }}>{item.message}</div>
                        <div style={{ display: "flex", gap: 6 }}>
                            {item.description && (
                                <button
                                    type="button"
                                    onClick={() => toggleExpand(item.id)}
                                    style={{
                                        border: "none",
                                        background: "transparent",
                                        fontSize: 12,
                                        color: "#1e293b",
                                        cursor: "pointer",
                                        padding: 0,
                                    }}
                                >
                                    {expandedIds.has(item.id) ? "Hide" : "Expand"}
                                </button>
                            )}
                            <button
                                type="button"
                                onClick={() => dismiss(item.id)}
                                style={{
                                    border: "none",
                                    background: "transparent",
                                    fontSize: 12,
                                    color: "#1e293b",
                                    cursor: "pointer",
                                    padding: 0,
                                }}
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
                                style={{
                                    marginTop: 6,
                                    border: "none",
                                    background: "#f1f5f9",
                                    color: "#0f172a",
                                    fontSize: 12,
                                    borderRadius: 6,
                                    padding: "4px 8px",
                                    cursor: "pointer",
                                }}
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