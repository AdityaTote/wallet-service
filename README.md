# Wallet Service

A digital wallet microservice implementing double-entry bookkeeping over PostgreSQL. Exposes a REST API for user registration, wallet top-up, spend, and balance queries.

## What This Is

- Go HTTP service (`chi` router, `pgx` driver, `sqlc` generated queries)
- PostgreSQL-backed ledger where balance = `SUM(ledger.amount)` per wallet
- Every mutation (topup/spend) creates two ledger entries that net to zero
- Concurrency handled via `SELECT ... FOR UPDATE` row locking within database transactions
- Client-supplied `txn_id` on every mutation for idempotent retries
- JWT authentication (HS256, 24h TTL)

## Quick Start

```bash
cp .env.example .env
docker compose up --build
```

This starts PostgreSQL, runs migrations, seeds test data, and starts the service on `http://localhost:8080`.

Pre-seeded test accounts: `alice`/`password123`, `bob`/`password456`, `charlie`/`password789`.

```bash
# Sign in
curl -s -X POST http://localhost:8080/api/auth/signin \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"password123"}' | jq .

# Use the returned access_token for wallet operations
curl -s http://localhost:8080/api/wallet/balance \
  -H "Authorization: Bearer <access_token>" | jq .
```

## Documentation

| Document | Path |
|---|---|
| Local setup & configuration | [docs/setup/local.md](docs/setup/local.md) |
| Architecture overview | [docs/architecture/overview.md](docs/architecture/overview.md) |
| Database schema & data model | [docs/architecture/database.md](docs/architecture/database.md) |
| API reference | [docs/api/reference.md](docs/api/reference.md) |
| Development workflow | [docs/development/workflow.md](docs/development/workflow.md) |
| Docker deployment | [docs/deployment/docker.md](docs/deployment/docker.md) |
| Operations & observability | [docs/operations/observability.md](docs/operations/observability.md) |
| Security & auth model | [docs/security/auth-model.md](docs/security/auth-model.md) |
| Troubleshooting | [docs/troubleshooting/common-issues.md](docs/troubleshooting/common-issues.md) |

### Decision Records

| ADR | Topic |
|---|---|
| [ADR-001](docs/decisions/001-technology-choices.md) | Technology choices |
| [ADR-002](docs/decisions/002-double-entry-ledger.md) | Double-entry ledger model |
| [ADR-003](docs/decisions/003-concurrency-strategy.md) | Concurrency strategy |

## Project Layout

```
cmd/
  seed/                 Standalone seed binary
  wallet-service/       Application entrypoint
docker/
  seed.dockerfile       Multi-stage build for seed
  wallet.dockerfile     Multi-stage build for service
internal/
  config/               Config loading (koanf, .env + env vars)
  database/             pgxpool connection
  handler/              HTTP handlers
  lib/utils/            JWT, bcrypt, JSON response helpers
  middleware/           Bearer token auth middleware
  models/               Request/response types, error types
  repository/           sqlc-generated queries + transaction wrapper
    queries/            Raw SQL source for sqlc
  router/               chi route definitions
  server/               HTTP server lifecycle
  service/              Business logic
  validations/          Input validation (go-playground/validator)
migrations/             golang-migrate SQL files
```
