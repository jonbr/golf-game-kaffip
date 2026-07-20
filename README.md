# Golf Game Kaffip

A Go API for scoring 2v2 team golf matches, supporting both handicap-adjusted
and gross-scoring formats, with live point-based match scoring across three
categories per hole.

## Features

- **Two game variants**
  - **Gross** — no per-hole handicap; the team with the higher combined
    handicap starts with a one-time points surplus.
  - **Net** — full per-hole handicap stroke allocation, except birdies and
    eagles, which always count on gross score.
- **Per-hole scoring across three categories**
  - Lowest individual score (1 point)
  - Lowest team accumulative score (1 point)
  - Birdie (1 point) / eagle (2 points) bonuses, always gross
- **Running match score** as a signed lead — the trailing team always
  shows 0, the leader shows the margin.
- **Corrections at any time** — any previously scored hole can be
  re-submitted, and the match score is fully replayed from scratch so
  corrections always ripple through correctly.
- **Course data snapshotting** — course and hole data (par, handicap
  index) is fetched once from an external API at game creation and
  stored locally, so scoring never depends on that API being available.

## Tech stack

- Go 1.26
- [chi](https://github.com/go-chi/chi) — HTTP router
- PostgreSQL 16, via [pgx](https://github.com/jackc/pgx)
- [golang-migrate](https://github.com/golang-migrate/migrate) — schema migrations
- [testcontainers-go](https://github.com/testcontainers/testcontainers-go) — integration tests against a real Postgres
- [tint](https://github.com/lmittmann/tint) — structured logging (`slog`) with readable dev output
- Docker / Docker Compose

## Architecture

The project follows a layered structure:

```
cmd/api              entrypoint
internal/
  api/                HTTP layer: handlers, middleware, DTOs, response/error helpers
  application/         services: orchestrate domain logic + repositories
  domain/              core business logic, no framework or DB dependencies
    game/              game aggregate, scoring rules, handicap math
    player/             player aggregate
    course/              course model
  infrastructure/
    postgres/            repository implementations
    external/opengolfapi/ external course data client
  bootstrap/            app wiring, config, DB connection, migrations
  config/               env-based configuration
  logging/               context-scoped logger
  testutil/              shared test helpers (testcontainers setup)
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

### Run locally (Go only, Postgres via Docker)

```bash
docker compose up -d postgres
make run-local
```

### Run the full stack in Docker

```bash
make dev
```

This builds the API image and starts both the API and Postgres via
Docker Compose. Use `make dev-detached` to run in the background, and
`make dev-logs` to follow the API's logs.

The API will be available at `http://localhost:8080`.

```bash
curl http://localhost:8080/health
```

## Testing

Tests are split into two tiers:

- **Unit tests** — pure domain logic, no I/O, run on every save.
- **Integration tests** — exercise the real repository layer against a
  disposable Postgres container (via testcontainers), gated behind the
  `integration` build tag so they never run as part of the fast everyday
  loop.

```bash
make test-unit          # fast, no Docker required beyond what's already running
make test-integration   # spins up a real Postgres container per test
make test-all           # both
```

Integration tests require Docker to be running.

## API overview

| Method | Path                                   | Description                           |
|--------|----------------------------------------|---------------------------------------|
| GET    | `/health`                              | Health check                          |
| POST   | `/players`                             | Create a player                       |
| GET    | `/players`                             | List players                          |
| GET    | `/players/{id}`                        | Get a player                          |
| PUT    | `/players/{id}`                        | Update a player                       |
| DELETE | `/players/{id}`                        | Soft-delete a player                  |
| POST   | `/games`                               | Create a game                         |
| GET    | `/games`                               | List games (optionally `?status=active\|finished`) |
| GET    | `/games/{id}`                          | Get full game state                   |
| PUT    | `/games/{id}/holes/{holeNumber}/score` | Submit or correct a hole's score      |
| POST   | `/games/{id}/finish`                   | Finish a game at its current hole     |

A player can only belong to one unfinished game at a time; finishing a
game frees its players to join another.

## Database

Schema migrations live in `internal/infrastructure/postgres/migrations`
and run automatically on startup. To reset the local database entirely:

```bash
docker compose down -v
docker compose up -d postgres
```
