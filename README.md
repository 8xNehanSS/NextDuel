# NextDuel

Real-time, head-to-head trivia duels. Two players are matched (or join a
private room by code), ready up, and race through multiple-choice rounds. The
first answer in a round cuts the opponent's clock; a round is won by being the
only player to answer correctly; most rounds won takes the match. Results feed
a match history and a leaderboard.

| Part | Stack |
| --- | --- |
| `nextduel-svr/` | Go 1.25, [nex-server](https://github.com/8xNehanSS/NEX-Server-Framework) (HTTP/WS framework), GORM + SQLite (pure Go, no cgo), gorilla/websocket |
| `nextduel-frontend/` | React 19, TypeScript, Vite, Zustand, React Router 7 |
| Deploy | Multi-stage Docker image, Caddy for TLS (`docker-compose.yml`) |

## Repository layout

The backend depends on the `nex-server` framework through a local `replace`
directive (`replace nex-server => ../../nex-server`), so it must be cloned
**next to** this repo:

```
Projects/
├── NextDuel/      ← this repo
└── nex-server/    ← git clone https://github.com/8xNehanSS/NEX-Server-Framework nex-server
```

## Local development

Prerequisites: Go 1.25+, Node 22+.

```bash
# 1. backend (http://localhost:8080)
cd nextduel-svr
go run .

# 2. frontend with hot reload (http://localhost:5173, proxies /api to :8080)
cd nextduel-frontend
npm install
npm run dev
```

To serve the production build from the Go server instead, run `npm run build`
and start the backend with `HTML_ROOT=../nextduel-frontend/dist`.

No account exists on a fresh database. Register through the UI, or set
`ADMIN_PASSWORD` before first start to create an `admin` user.

## Configuration

All settings are environment variables (a `.env` file in the working
directory is loaded automatically). See [`.env.example`](.env.example) for the
full list with defaults. The important ones:

| Variable | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | HTTP port |
| `DB_PATH` | `nextduel.db` | SQLite file (WAL mode) |
| `HTML_ROOT` | `./html` | Built frontend directory |
| `ADMIN_PASSWORD` | – | Seeds an `admin` account if none exists |
| `TRUST_PROXY` | `false` | Use `X-Forwarded-For` for rate limiting (only behind a proxy) |
| `OPENTDB_ENABLED` | `true` | Fetch questions from opentdb.com; falls back to the local bank |
| `GAME_ROUNDS` / `GAME_ROUND_SECONDS` / `GAME_FAST_DUEL_SECONDS` / `GAME_INTERMISSION_SECONDS` | `5` / `60` / `10` / `5` | Game tuning |

## Testing

```bash
cd nextduel-svr && go test ./...          # add -race on Linux/macOS (needs cgo)
cd nextduel-frontend && npm run lint && npm run build
```

The backend suite includes engine tests with millisecond timers, repository
tests against a real SQLite file, and an end-to-end test
(`main_test.go`) that registers two users and plays a full duel over real
websockets. CI (`.github/workflows/ci.yml`) runs all of this plus a Docker
build.

## Deployment

```bash
cp .env.example .env        # set DOMAIN, ADMIN_PASSWORD, ...
docker compose up -d --build
```

The image is built from the **parent** directory (it needs `nex-server`):

```bash
cd Projects
docker build -f NextDuel/Dockerfile -t nextduel:latest --build-arg VERSION=1.0.0 .
docker save -o nextduel.tar nextduel:latest   # export
docker load -i nextduel.tar                   # import on the host
```

The database (`/app/nextduel-svr/data`) and login sessions
(`/app/nextduel-svr/sessions`) live in named volumes. The container has a
built-in health check (`nextduel-svr healthcheck` → `GET /api/health`).

> **Upgrading an existing deployment:** earlier versions mounted
> `./nextduel.db` to a path the app never wrote to, so data lived only inside
> the container. Before recreating it, copy the database out:
> `docker cp nextduel:/app/nextduel-svr/nextduel.db ./nextduel.db`, then copy
> it into the new `nextduel_data` volume.

## API

REST (JSON, `{success, data}` / `{success: false, error}` envelope):

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| GET | `/api/health` | – | Liveness + DB check |
| POST | `/api/register` | – | Create account (rate limited) |
| POST | `/api/login` | – | Sets the session cookie (rate limited) |
| POST | `/api/logout` | ✓ | Clears the session |
| GET | `/api/user` | ✓ | Current user |
| GET | `/api/stats` | ✓ | Live server statistics |
| GET | `/api/history/duels?limit=&offset=` | ✓ | Your finished matches + stats |
| GET | `/api/leaderboard` | ✓ | Top players |
| GET | `/api/ws` | ✓ | Game websocket |

Websocket messages are `{"type": string, "data": any}`. The protocol is
defined once per side and must stay in sync:
[`nextduel-svr/utils/events.go`](nextduel-svr/utils/events.go) +
[`gamepkg/protocol.go`](nextduel-svr/gamepkg/protocol.go) ↔
[`src/utils/websocket/events.ts`](nextduel-frontend/src/utils/websocket/events.ts) +
[`src/types/protocol.ts`](nextduel-frontend/src/types/protocol.ts).
