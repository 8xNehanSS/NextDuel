/**
 * Returns `target` only if it is a same-origin path, preventing open
 * redirects via ?redirect=https://evil.example or //evil.example.
 */
export function safeRedirect(target: string | null | undefined, fallback = "/"): string {
    if (!target || !target.startsWith("/") || target.startsWith("//") || target.startsWith("/\\")) {
        return fallback;
    }
    return target;
}
