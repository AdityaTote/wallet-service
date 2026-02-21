# Local Setup

## Prerequisites

### Docker (recommended)

- Docker Engine 20+
- Docker Compose v2

### Without Docker

- Go 1.25+
- PostgreSQL 15+
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI
- [sqlc](https://sqlc.dev/) (only if modifying SQL queries)

## Configuration

All configuration is via environment variables. Copy the template:

```bash
cp .env.example .env
```

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server listen port |
| `DATABASE_HOST` | `localhost` | PostgreSQL host |
| `DATABASE_PORT` | `5432` | PostgreSQL port |
| `DATABASE_USER` | `postgres` | PostgreSQL user |
| `DATABASE_PASSWORD` | `postgres` | PostgreSQL password |
| `DATABASE_NAME` | `postgres` | PostgreSQL database name |
| `JWT_SECRET` | (none) | HMAC-SHA256 signing key for JWT tokens |

The application loads config in this order (later sources override earlier):
1. `.env` file (if it exists on disk)
2. Environment variables

In Docker Compose, the `env_file: .env` directive injects these into containers. The Docker Compose file also hardcodes the PostgreSQL container's credentials (`postgres`/`postgres`/`postgres`), so the `.env.example` defaults work out of the box.

## Running with Docker Compose

```bash
docker compose up --build
```

Startup order (enforced by `depends_on`):

1. `postgres` — starts, waits for `pg_isready` health check
2. `migrate` — runs `UP` migrations against the healthy database
3. `seed-service` — inserts seed data, then exits
4. `wallet-service` — starts HTTP server on port 8080

To tear down and remove volumes:

```bash
docker compose down -v
```

## Running Locally (without Docker)

You need a running PostgreSQL instance.

```bash
# 1. Configure
cp .env.example .env
# Edit .env: set DATABASE_HOST, DATABASE_PORT, etc. to match your PostgreSQL

# 2. Run migrations
make migrate-up

# 3. Seed the database
make seed

# 4. Start the service
make run
```

The `Makefile` reads `.env` via `include .env` / `export`.

### Makefile Targets

| Target | Command | Description |
|---|---|---|
| `migrate-up` | `migrate -path migrations -database "..." up` | Apply all pending migrations |
| `seed` | `go run cmd/seed/main.go` | Insert seed data (idempotent) |
| `run` | `go run cmd/wallet-service/main.go` | Start the service |
| `sqlc-generate` | `sqlc generate` | Regenerate Go code from SQL queries |

## Seed Data

The seed script (`cmd/seed/main.go`) creates:

| Entity | Details |
|---|---|
| Asset | `UC` (Universal Credits) |
| System wallet | `SYSTEM` owner, 1,000,000,000 UC initial balance |
| `alice` | password: `password123`, 10,000 UC |
| `bob` | password: `password456`, 5,000 UC |
| `charlie` | password: `password789`, 20,000 UC |

The script is idempotent: it checks for existing assets/users before inserting. To re-seed, manually delete from `ledgers`, `transactions`, `wallets`, `users`, `assets` in that order (foreign key constraints).

## Verifying the Setup

```bash
# Health check
curl http://localhost:8080/api/health

# Expected: {"success":true,"message":"service is healthy","data":{"status":"healthy",...}}
```

If the database check shows `"unhealthy"`, the service cannot reach PostgreSQL.
