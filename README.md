# Golf Game Kaffip

A Go API for scoring golf games between friends — head-to-head match play,
2v2 team points, 4-player Wolf, and Ryder-Cup-style Cups that group several
finished matches of one format into one aggregate score.

## Features

- **Three ways to play**
  - **Match play** — classic 1v1, hole-by-hole. Lower score wins the hole;
    ties halve it. Result reported in classic terms (`"3&2"`, `"2 up"`,
    `"All Square"`, `"Halved"`).
  - **Team points** — 2v2, points-based, scored across three categories per
    hole: lowest individual score, lowest team accumulative score, and
    birdie/eagle bonuses (always gross, regardless of variant). The
    running score is a signed lead — the trailing team always shows 0.
  - **Wolf** — 4 players, a rotating "Wolf" seat each hole. The Wolf either
    partners with one other player or goes it alone against the other
    three; the winning side (Wolf's side vs. the field) splits the hole's
    points, tied holes award nothing. Always net-scored today (no gross
    option yet).
- **Cups** — group several already-playable matches into one Ryder-Cup-style
  event. A Cup is type-homogeneous: every match inside one Cup is the same
  format (all Match Play, all Team Points, or all Wolf). Match Play and
  Team Points matches are each worth 1 point toward the Cup once finished
  (0.5/0.5 if tied); a Wolf match's point goes to whichever Cup side its
  best individual performer belongs to, with ties broken by comparing
  combined side totals. The Cup's aggregate score is never stored — always
  recomputed live from each match's current state, so it can't drift out
  of sync.
- **Two scoring variants**, orthogonal to game type
  - **Gross** — no per-hole handicap. Team Points games start with a
    one-time points surplus for the team with the higher combined handicap.
  - **Net** — full per-hole handicap stroke allocation. Birdies/eagles
    always count on gross score regardless of variant.
- **Corrections at any time** — any previously scored hole can be
  re-submitted, and the affected game's score is fully replayed from
  scratch so corrections always ripple through correctly, even after a
  game is "finished" — the app doubles as a bookkeeping tool.
- **Course data snapshotting** — course and hole data (name, par, handicap
  index) is fetched once from an external API (OpenGolfAPI) at game
  creation and stored locally, so scoring never depends on that API being
  reachable afterward. Course search is also available for picking a
  course by name/city/state.

## Tech stack

- Go 1.26
- [chi](https://github.com/go-chi/chi) — HTTP router
- PostgreSQL 16, via [pgx](https://github.com/jackc/pgx)
- [golang-migrate](https://github.com/golang-migrate/migrate) — schema migrations
- [testcontainers-go](https://github.com/testcontainers/testcontainers-go) — integration tests against a real Postgres
- [tint](https://github.com/lmittmann/tint) — structured logging (`slog`) with readable dev output
- Docker / Docker Compose

## Architecture

```
cmd/
  api/                 main API server entrypoint
  migrate/              standalone migration runner (no server startup)
internal/
  api/                  HTTP layer: per-format handlers, middleware, DTOs, response/error helpers
  application/          services: orchestrate domain logic + repositories, decoupled per format
                         (Match Play, Team Points, Wolf, and their three Cup services)
  domain/               core business logic, no framework or DB dependencies
    game/                shared Game aggregate (Match Play + Team Points) and handicap math
      matchplay/           Match Play hole-result scoring
      teampoints/          Team Points hole-result scoring
    wolf/                Wolf aggregate: rotating Wolf seat, its own scoring + standings
    cup/                 Cup aggregate: groups finished matches of one format into a score
    player/               player aggregate
    course/                 course model
  infrastructure/
    postgres/             repository implementations (game/, wolf/, cup/, player/)
    external/opengolfapi/ external course data client + search
  bootstrap/             app wiring, config, DB connection, migrations
  config/                 env-based configuration
  logging/                 context-scoped logger
  testutil/                shared test helpers (testcontainers setup)
```

Domain logic has no knowledge of HTTP or SQL — it operates purely on
in-memory structs, which keeps the scoring rules easy to test and safe to
reason about independently of persistence or transport concerns.

## Getting started

### Prerequisites

- Go 1.26+
- Docker (for Postgres, and for running the full stack)
- An [OpenGolfAPI](https://opengolfapi.org) API key

### Environment variables

Create a `.env` file in the project root:

```
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/golf_game?sslmode=disable
OPENGOLF_API_KEY=your-api-key-here
```

Note: `DATABASE_URL` should use `localhost` when running the API locally
against a Dockerized Postgres, or the Compose service name (`postgres`)
when running the whole stack together via Compose. Don't run both at once —
they'll fight over port 8080.

### Recommended local dev setup: Postgres in Docker, API native

```bash
docker compose up -d postgres
go run ./cmd/migrate      # apply any pending migrations
go run cmd/api/main.go    # or: make run-local
```

### Run the full stack in Docker

```bash
make dev
```

Builds the API image and starts both the API and Postgres via Docker
Compose. Use `make dev-detached` to run in the background, and
`make dev-logs` to follow the API's logs. Don't run this alongside a
locally-running `go run cmd/api/main.go` — both bind port 8080.

The API will be available at `http://localhost:8080`.

```bash
curl http://localhost:8080/health
```

`/health` also checks database connectivity and returns `503` if the DB is
unreachable, not just whether the process is alive.

### Running migrations independently

`cmd/migrate` runs migrations as a standalone step, independent of
starting the API server:

```bash
go run ./cmd/migrate
```

Migration files live in `internal/infrastructure/postgres/migrations` and
are always additive (new numbered files, e.g. `0002_add_cups`), never
edited once applied to any database with real data.

## Testing

Tests are split into two tiers:

- **Unit tests** — pure domain logic, no I/O, run on every save.
- **Integration tests** — exercise the real repository layer against a
  disposable Postgres container (via testcontainers), gated behind the
  `integration` build tag so they never run as part of the fast everyday
  loop.

```bash
make test-unit          # fast, no Docker container spin-up needed
make test-integration   # spins up a real Postgres container per test
make test-all           # both
```

Integration tests require Docker to be running.

For nicer local output (colorized, one line per test), optionally install
[gotestsum](https://github.com/gotest.tools/gotestsum) and run it directly —
it's a personal dev-loop tool, not part of the canonical Makefile targets:

```bash
gotestsum --format testname ./...
gotestsum --format testname -- -tags=integration ./...
```

## API overview

### Players

| Method | Path | Description |
|--------|------|--------------|
| POST   | `/players` | Create a player (name, email, handicap — email is mandatory and unique) |
| GET    | `/players` | List players |
| GET    | `/players/{id}` | Get a player |
| PUT    | `/players/{id}` | Update a player |
| DELETE | `/players/{id}` | Soft-delete a player |

A player can only belong to one unfinished game (or Cup match) at a
time; finishing frees them.

### Courses

| Method | Path | Description |
|--------|------|--------------|
| GET    | `/courses/search?q=...` | Search courses by name/city/state via OpenGolfAPI |

### Games (standalone Match Play, Team Points, or Wolf)

| Method | Path | Description |
|--------|------|--------------|
| POST   | `/games/match_play` | Create a 1v1 match play game |
| POST   | `/games/team_points` | Create a 2v2 team points game |
| POST   | `/games/wolf` | Create a 4-player Wolf game |
| GET    | `/games?status=active\|finished` | List games (lightweight summaries) |
| GET    | `/games/{id}` | Get full game state (generic; Wolf also has its own `GetWolfPlay` for full detail) |
| PUT    | `/games/{id}/holes/{holeNumber}/score` | Submit or correct a hole's score |
| POST   | `/games/{id}/finish` | Finish a game at its current hole |

### Cups (grouped matches, Ryder Cup style)

| Method | Path | Description |
|--------|------|--------------|
| POST   | `/cups/match_play` | Create a Match Play Cup: N 1v1 matches, plus a two-sided player roster |
| POST   | `/cups/team_points` | Create a Team Points Cup: N 2v2 matches |
| POST   | `/cups/wolf` | Create a Wolf Cup: N standalone 4-player Wolf games |
| GET    | `/cups/{id}` | Full Cup: roster, aggregate score, all matches |
| POST   | `/cups/{id}/finish` | Mark the whole Cup finished |

A Cup is type-homogeneous — every match inside one Cup is the same format,
unlike a mixed grouping. Individual matches inside a Cup are scored and
finished using the **same** per-format `/games/{id}/holes/{holeNumber}/score`
and `/games/{id}/finish` endpoints above, by their own `game_id` — there's
no Cup-specific scoring endpoint. Each Match Play/Team Points match is worth
1 point toward the Cup once finished (0.5/0.5 if tied); a Wolf match's point
goes to whichever Cup side its best individual performer belongs to (tied
by comparing combined side totals). The Cup's aggregate score is always
derived fresh from current match state, never stored.

### Health

| Method | Path | Description |
|--------|------|--------------|
| GET    | `/health` | Health check, including DB connectivity |

## Database

Schema migrations live in `internal/infrastructure/postgres/migrations`.
To reset the local database entirely during development:

```bash
docker compose down -v
docker compose up -d postgres
go run ./cmd/migrate
```