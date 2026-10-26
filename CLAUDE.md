# CLAUDE.md

Guidance for working in this repository. See README.md for user-facing setup.

## What this is

NextDuel: real-time 1v1 duels in several games (trivia, math sprint, word
scramble, rock-paper-scissors). Go backend (`nextduel-svr/`) + React SPA
(`nextduel-frontend/`). Players pick a game, queue or join a private room by
6-char code, ready up, then play `GAME_ROUNDS` rounds over a websocket.

## Commands

```bash
# from the repo root (see package.json / scripts/setup.mjs)
npm run setup                     # clones ../nex-server if missing, writes nextduel-svr/.env, installs deps
npm run dev                       # backend + Vite together; backend PORT comes from nextduel-svr/.env
npm test                          # everything CI checks except -race and gofmt

# backend (run from nextduel-svr/)
go build ./... && go vet ./...
go test ./... -count=1           # -race needs cgo; not available on the Windows dev box, CI runs it
gofmt -l .                        # CI fails on unformatted files
go run .                          # :8080

# frontend (run from nextduel-frontend/)
npm run lint                      # must be clean, CI enforces it
npm run build                     # tsc -b + vite build
npm run dev                       # 127.0.0.1:5173, proxies /api + ws to PORT from ../nextduel-svr/.env
```

## External dependency: nex-server

`go.mod` has `replace nex-server => ../../nex-server`. The framework is a
**separate, private repo** (github.com/8xNehanSS/NEX-Server-Framework) checked out as a
sibling of this repo. Do not edit it from here without asking. Things it
provides and constraints it imposes:

- `server.Context`: `ctx.UserID()` is the **username**, used as the player ID everywhere.
- Responses: `ctx.Success/Created` → `{success:true,data}`, `ctx.Error` → `{success:false,error}`.
  The frontend `apiclient` unwraps this envelope.
- Sessions are files in `./sessions` (relative to cwd), 24h expiry, cookie named `COOKIE_KEY`.
- One goroutine per websocket runs `readLoop`, calling our handler per message.
  The handler returning an error closes the connection, so ours never does.
- Known framework issues (not fixable here): hard-coded AES key in
  `utils/crypt.go`, session cookie lacks `Secure`, `CheckOrigin` accepts all
  (we wrap `/api/ws` with `middleware.CheckOrigin`), `Start()` has no
  timeouts (we run `s.Router()` on our own `http.Server` in `main.go`).

## Backend architecture (`nextduel-svr/`)

```
main.go          wiring: newApp() builds routes; run() = http.Server + graceful shutdown;
                 `nextduel-svr healthcheck` subcommand for Docker
config/          NextDuel env config (framework reads PORT/COOKIE_KEY/HTML_ROOT itself)
database/        Open (SQLite WAL), Migrate (AutoMigrate), seeding
models/          GORM models. Players (game stats) is separate from User (auth), keyed by username
gamepkg/         game engine, no DB imports
  manager.go     GameManager: sessions (per *websocket.Conn), queue, rooms, playerRooms
  room.go        Room: one duel; run() goroutine drives rounds/timers (game-agnostic)
  modes.go       Mode interface + registry + shared resolvers (resolveExclusive/resolveFastest)
  mode_*.go      one file per game: trivia, math, scramble, rps
  player.go      Player: one websocket session with a single writer goroutine
  recorder.go    Recorder interface (persistence port) + noop
  protocol.go    wire payload structs
  errors.go      GameError{Code,Message} sent to clients as "error" events
questions/       Provider interface; OpenTDB client; Bank (quest table); Fallback chain; seed.json
repository/      MatchRecorder (implements gamepkg.Recorder), History, Leaderboard queries
middleware/      security headers, recover, access log, rate limiter, ws origin check, SPA handler
routes/          thin HTTP/WS handlers; game-events/handler.go is the WS dispatch table
utils/           GameAction + status enums, room code generator
```

### Concurrency rules (read before touching gamepkg)

- **Never write to a websocket directly.** Use `Player.Send`, which queues to
  the player's writer goroutine. gorilla/websocket panics on concurrent writes.
- `GameManager.mu` guards its maps; `Room.mu` guards room state.
- **Lock order is manager → room.** `AddPlayer`, `SetReady`, `SubmitAnswer` and
  `State` may be called with `GameManager.mu` held. `Leave` and `Abort` may close
  the room, which calls `onClosed` → manager lock, so they **must not** be called
  with `GameManager.mu` held. `Room` never calls the manager while holding `Room.mu`.
- Methods named `*Locked` require the caller to hold the relevant mutex.
- Room identity checks use the `*Player` pointer, not the ID, so a second
  browser tab (same user, different conn) cannot act in or forfeit the first tab's game.
- `Recorder` methods run under `Room.mu`. They must not call back into the room,
  and they resolve player IDs **before** opening a transaction (SQLite has one writer).

### Game modes

Every game is the same duel shape, so `Room` stays game-agnostic and a
`Mode` supplies the content and scoring:

- `Mode.Rounds(ctx, n)` returns n prompts (`questions.Question`: text, choices, optional answer).
- `Mode.Resolve(q, answers)` returns who scored (+1), the round winner, the revealed answer and a summary.
  It runs under `Room.mu`; keep it pure.
- `ModeInfo.RoundDuration` caps the round clock (config can only shorten it further).

| ID | Rounds | Round winner |
| --- | --- | --- |
| `trivia` | questions provider (OpenTDB → bank) | only player correct (`resolveExclusive`) |
| `math` | generated, difficulty ramps per round | first correct answer (`resolveFastest`) |
| `scramble` | built-in word list; decoys are letter-swaps | first correct answer |
| `rps` | fixed Rock/Paper/Scissors | classic rules; not picking loses; tie = none |

"First" uses `RoundAnswer.Order` (assigned under the lock), not wall-clock
elapsed time: Windows timer resolution makes near-simultaneous answers equal.

**Adding a game:** implement `Mode` in `gamepkg/mode_<id>.go`, add the ID to
`utils.GameMode`, register it in `newModes` (main.go), add the TS union member
in `src/types/protocol.ts` and an entry in `src/games.ts` (name, icon, rules).
Add special rendering in `pages/game/page.tsx` (`Prompt`) only if needed.
Queues, rooms, history, per-game leaderboards and persistence need no changes.

### Rules shared by every mode (encoded in room.go)

- Answers must exactly match one of the offered choices.
- The first answer shortens the remaining clock to `FastDuelDuration`. Answers stay hidden from the opponent until the round resolves.
- Match winner = more rounds won (tie = draw). Leaving mid-match = forfeit (match `abandoned`, opponent wins).
- Matchmaking queues are per mode; a player can be in only one queue or room at a time.
- `join_queue` / `create_room` take `{mode}`; a missing mode means `trivia` (older clients).
- `matches.mode` stores the mode; `database.Migrate` rewrites legacy `classic` rows to `trivia`.
- Match statuses: `active`, `completed`, `abandoned` (forfeit, counts in stats),
  `aborted` (no result, e.g. shutdown; hidden from history and stats).
- A DB match row is created only when the game actually starts.

### Questions

`Fallback{Primary: OpenTDB, Secondary: Bank}`. OpenTDB rate-limits to ~1
request per 5s per IP, so under load the bank serves most games. Fetched
questions are cached into the bank. OpenTDB text is HTML-entity encoded and is
decoded in `opentdb.go`.

## Websocket protocol

Envelope `{"type", "data"}` both ways. **Keep these four files in sync** when
adding or changing a message:

- `nextduel-svr/utils/events.go`: `GameAction` names
- `nextduel-svr/gamepkg/protocol.go`: payload structs
- `nextduel-frontend/src/utils/websocket/events.ts`: names + `ServerEvents`/`ClientEvents` type maps
- `nextduel-frontend/src/types/protocol.ts`: payload types

The server pushes full `RoomState` snapshots (`room_state`, `match_found`,
`room_created`, `room_joined`) rather than incremental membership events.

## Frontend architecture (`nextduel-frontend/src/`)

- `utils/websocket/socket.ts`: singleton `socketManager`; lazy connect on first `send`,
  25s heartbeat, resets game state and notifies on unexpected close.
- `utils/websocket/handler.ts`: maps server events to `gameStore` actions. Keep logic here, not in components.
- `store/gameStore.ts`: Zustand. `phase` drives the game page:
  `idle → queued → lobby → starting → question ⇄ round_result → game_over`.
- `store/userStore.ts`: auth state. `authChecked` gates route guards.
- `utils/apiclient.ts`: fetch wrapper; throws `HttpError` with the server's message.
- `utils/navigationService.ts`: lets non-component code navigate.
- Styling: plain CSS design tokens in `index.css` + component classes in `App.css`.
  Tailwind is installed but not used by current pages; prefer the existing classes.

## Testing conventions

- `gamepkg` tests use `fakeConn`, `staticProvider` (answer is always "A") and
  ms-scale `testSettings()`. Wait on messages with `conn.expect(t, action)`, never `time.Sleep`.
- Repository/questions tests use a temp-file SQLite DB (`t.TempDir()`).
- `main_test.go` is the end-to-end test through real HTTP + websockets; extend it for new flows.

## Gotchas

- Windows dev box: files may be CRLF on checkout; `.gitattributes` normalises. Run `gofmt -w` before committing.
- Releases: pushing any tag runs `.github/workflows/release.yml`, which calls `ci.yml` (reusable via `workflow_call`, `docker: false`) and then publishes a `docker save` tarball (`nextduel-<tag>.tar.gz` + `.sha256`) to a GitHub Release and as a workflow artifact. The tag becomes `main.version`, which `/api/health` reports.
- CI checks out the private nex-server repo with the `NEX_SERVER_DEPLOY_KEY` secret (read-only deploy key; see README "CI"). Any new job that builds the backend needs that checkout step too.
- The Docker build context is the **parent** directory (needs `nex-server/`);
  ignore rules live in `Dockerfile.dockerignore`.
- `nextduel.db` and `sessions/` in `nextduel-svr/` are local dev data (gitignored).
- On the main dev machine port 8080 is taken by a system process, so the local
  backend runs on 8081 (set in the gitignored `nextduel-svr/.env`).
- Vite binds `127.0.0.1` on purpose; `localhost` resolved to an unreachable `::1` listener on Windows.
