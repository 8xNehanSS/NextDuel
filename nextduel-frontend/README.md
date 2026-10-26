# nextduel-frontend

React 19 + TypeScript + Vite single-page app for NextDuel. See the
[repository README](../README.md) for setup.

```bash
npm install
npm run dev      # http://localhost:5173, proxies /api (incl. websocket) to :8080
npm run lint
npm run build    # outputs dist/, served by the Go server via HTML_ROOT
```

Set `VITE_BACKEND_URL` to proxy to a backend other than `http://localhost:8080`.
