# nextduel-frontend

React 19 + TypeScript + Vite single-page app for NextDuel. See the
[repository README](../README.md) for setup.

```bash
npm install
npm run dev      # http://localhost:5173, proxies /api (incl. websocket) to the backend
npm run lint
npm run build    # outputs dist/, served by the Go server via HTML_ROOT
```

The backend port is read from `../nextduel-svr/.env` (`PORT`). Set
`VITE_BACKEND_URL` to proxy somewhere else entirely. Usually you want
`npm run dev` from the repo root, which starts the backend too.
