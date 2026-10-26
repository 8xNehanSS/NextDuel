import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// The Go server serves the built app in production. In development, Vite
// proxies /api (including the /api/ws websocket) to it so cookies stay
// same-origin. The backend port comes from nextduel-svr/.env, so both sides
// always agree; VITE_BACKEND_URL overrides it entirely.
export default defineConfig(({ mode }) => {
    const backendEnv = loadEnv(mode, "../nextduel-svr", "");
    const backend = process.env.VITE_BACKEND_URL ?? `http://127.0.0.1:${backendEnv.PORT || "8080"}`;

    return {
        plugins: [react(), tailwindcss()],
        server: {
            // Bind IPv4 explicitly: "localhost" can resolve to ::1 only on
            // Windows, which some setups cannot reach. Browsers still open
            // http://localhost:5173 fine via IPv4 fallback.
            host: "127.0.0.1",
            proxy: {
                "/api": { target: backend, ws: true },
            },
        },
    };
});
