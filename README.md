# NextDuel

Real-time, head-to-head duels. Pick a game, get matched (or join a private
room by code), ready up, and race through the rounds. The first answer in a
round cuts the opponent's clock, and most rounds won takes the match. Results
feed a match history and per-game leaderboards.

| Game | How a round is won |
| --- | --- |
| 🧠 Trivia Duel | Be the only player to answer a general-knowledge question correctly |
| ➗ Math Sprint | First correct answer to a mental-arithmetic problem (gets harder each round) |
| 🔤 Word Scramble | First to pick the real word from same-letter decoys |
| ✊ Rock Paper Scissors | Classic rules; picks stay hidden until the reveal |

Adding another game is a single Go file plus a frontend registry entry; see
"Game modes" in [CLAUDE.md](CLAUDE.md).

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

`NEX-Server-Framework` is a **private** repository, so you need read access to
it. `npm run setup` clones it with your normal git credentials.

## Local development

Prerequisites: Go 1.25+, Node 20+ (22 recommended), git.

```bash
npm run setup   # one-time: clones nex-server if missing, creates nextduel-svr/.env
                # (picks a free port if 8080 is taken), installs all dependencies
npm run dev     # backend + frontend with hot reload -> http://localhost:5173
```

All commands run from the repo root:

| Command | What it does |
| --- | --- |
| `npm run setup` | Safe to re-run. Adds new settings to an existing `.env` without overwriting yours |
| `npm run dev` | Go backend + Vite dev server side by side; open the URL Vite prints |
| `npm start` | Builds the frontend, then serves everything from Go at `http://localhost:<PORT>` |
| `npm run build` | Frontend build + backend binary in `nextduel-svr/bin/` |
| `npm test` | Backend vet + tests, frontend lint + typecheck/build |

Local settings live in `nextduel-svr/.env` (template:
[`nextduel-svr/.env.example`](nextduel-svr/.env.example)). The Vite proxy
reads `PORT` from that file, so changing the backend port needs no other edits.
The local template uses shorter round timers for faster manual testing.

No account exists on a fresh database. Register through the UI, or set
`ADMIN_PASSWORD` in `.env` before first start to create an `admin` user. To
play a duel by yourself, sign in as a second user in a private window.

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

## CI

The backend and Docker jobs must check out the private `NEX-Server-Framework`
repo. They use a read-only **deploy key**, which grants access to that one
repository only. One-time setup:

1. Generate a key pair with no passphrase. Run this outside the repo so the key is never committed:
   ```bash
   ssh-keygen -t ed25519 -N "" -C "nextduel-ci" -f nextduel-ci-key
   ```
2. **NEX-Server-Framework** → Settings → Deploy keys → *Add deploy key*:
   paste `nextduel-ci-key.pub` and leave *Allow write access* unchecked.
3. **NextDuel** → Settings → Secrets and variables → Actions → *New repository
   secret*: name `NEX_SERVER_DEPLOY_KEY`, value = the full contents of
   `nextduel-ci-key` (the private key, including the `BEGIN`/`END` lines).
4. Delete both local key files and re-run the failed workflow.

If the secret is missing, the backend job fails immediately with an explanatory
error instead of a confusing checkout failure. Pull requests from forks never
receive secrets, so their backend job fails the same way; the frontend job
still runs.

## Deployment

```bash
cp .env.example .env        # set DOMAIN, ADMIN_PASSWORD, ...
docker compose up -d --build
```

### Releases (recommended)

Pushing a tag builds and publishes the image automatically
(`.github/workflows/release.yml`):

```bash
git tag v1.2.0
git push origin v1.2.0
```

The workflow runs the full CI checks, builds the image (tagged `nextduel:v1.2.0`
and `nextduel:latest`, with the version baked into `/api/health`), smoke-tests
it, and attaches `nextduel-v1.2.0.tar.gz` plus a `.sha256` checksum to a new
**GitHub Release**. The same files are also kept as a workflow artifact for 90
days. On the server:

```bash
sha256sum -c nextduel-v1.2.0.tar.gz.sha256
gunzip -c nextduel-v1.2.0.tar.gz | docker load
docker compose up -d          # without --build, so it uses the loaded image
```

### Building by hand

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
| GET | `/api/history/duels?limit=&offset=&mode=` | ✓ | Your finished matches + stats (optionally one game) |
| GET | `/api/leaderboard?mode=` | ✓ | Top players overall, or for one game |
| GET | `/api/ws` | ✓ | Game websocket |

Websocket messages are `{"type": string, "data": any}`. The protocol is
defined once per side and must stay in sync:
[`nextduel-svr/utils/events.go`](nextduel-svr/utils/events.go) +
[`gamepkg/protocol.go`](nextduel-svr/gamepkg/protocol.go) ↔
[`src/utils/websocket/events.ts`](nextduel-frontend/src/utils/websocket/events.ts) +
[`src/types/protocol.ts`](nextduel-frontend/src/types/protocol.ts).
