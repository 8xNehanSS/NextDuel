import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// The Go server serves the built app in production. In development, Vite
// proxies /api (including the /api/ws websocket) to it so cookies stay
// same-origin.
const backend = process.env.VITE_BACKEND_URL ?? "http://localhost:8080";

export default defineConfig({
    plugins: [react(), tailwindcss()],
    server: {
        proxy: {
            "/api": { target: backend, ws: true },
        },
    },
});
