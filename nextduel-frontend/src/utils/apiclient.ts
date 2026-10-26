type HttpMethod = "GET" | "POST" | "PUT" | "DELETE" | "PATCH";

interface RequestOptions<TBody = unknown> {
    method?: HttpMethod;
    body?: TBody;
    headers?: Record<string, string>;
    query?: Record<string, string | number | boolean | undefined>;
    signal?: AbortSignal;
}

/** Envelope used by nex-server's ctx.Success / ctx.Created / ctx.Error. */
interface Envelope<T> {
    success: boolean;
    data?: T;
    error?: string;
}

export class HttpError extends Error {
    readonly status: number;
    readonly data: unknown;

    constructor(status: number, message: string, data: unknown) {
        super(message);
        this.name = "HttpError";
        this.status = status;
        this.data = data;
    }
}

const BASE_URL = "/api";

function buildQuery(query?: RequestOptions["query"]) {
    if (!query) return "";
    const params = new URLSearchParams();
    for (const [key, value] of Object.entries(query)) {
        if (value !== undefined && value !== null) {
            params.append(key, String(value));
        }
    }
    const qs = params.toString();
    return qs ? `?${qs}` : "";
}

function isEnvelope(value: unknown): value is Envelope<unknown> {
    return typeof value === "object" && value !== null && "success" in value;
}

/**
 * Performs a request and returns the parsed body. Responses wrapped in the
 * server's {success, data} envelope are unwrapped to `data`. Non-2xx responses
 * throw HttpError carrying the server's message.
 */
async function request<TResponse, TBody = unknown>(endpoint: string, options: RequestOptions<TBody> = {}): Promise<TResponse> {
    const { method = "GET", body, headers = {}, query, signal } = options;

    const res = await fetch(`${BASE_URL}${endpoint}${buildQuery(query)}`, {
        method,
        credentials: "same-origin",
        headers: {
            Accept: "application/json",
            ...(body !== undefined ? { "Content-Type": "application/json" } : {}),
            ...headers,
        },
        body: body !== undefined ? JSON.stringify(body) : undefined,
        signal,
    });

    let data: unknown = null;
    if (res.status !== 204) {
        try {
            data = await res.json();
        } catch {
            data = null;
        }
    }

    if (!res.ok) {
        const message = isEnvelope(data) && data.error ? data.error : res.statusText || "Request failed";
        throw new HttpError(res.status, message, data);
    }

    return (isEnvelope(data) ? data.data : data) as TResponse;
}

export const api = {
    get: <T>(url: string, options?: Omit<RequestOptions, "body" | "method">) => request<T>(url, { ...options, method: "GET" }),

    post: <T, B = unknown>(url: string, body?: B, options?: RequestOptions<B>) =>
        request<T, B>(url, { ...options, body, method: "POST" }),

    put: <T, B = unknown>(url: string, body?: B, options?: RequestOptions<B>) =>
        request<T, B>(url, { ...options, body, method: "PUT" }),

    patch: <T, B = unknown>(url: string, body?: B, options?: RequestOptions<B>) =>
        request<T, B>(url, { ...options, body, method: "PATCH" }),

    delete: <T>(url: string, options?: RequestOptions) => request<T>(url, { ...options, method: "DELETE" }),
};

export function errorMessage(error: unknown, fallback = "Something went wrong"): string {
    if (error instanceof HttpError || error instanceof Error) {
        return error.message || fallback;
    }
    return fallback;
}
