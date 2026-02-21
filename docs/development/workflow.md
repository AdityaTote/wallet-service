# Development Workflow

## Building

```bash
# Build the service binary
go build -o bin/wallet-service ./cmd/wallet-service

# Build the seed binary
go build -o bin/seed ./cmd/seed
```

No build tags or special flags are required.

## Running

See [setup/local.md](../setup/local.md) for full setup. Quick summary:

```bash
make run          # go run cmd/wallet-service/main.go
make seed         # go run cmd/seed/main.go
make migrate-up   # apply migrations
```

## Database Migrations

Migrations use [golang-migrate](https://github.com/golang-migrate/migrate). Files live in `migrations/`.

```bash
# Apply all pending
make migrate-up

# Create a new migration (manual)
migrate create -ext sql -dir migrations -seq <description>
```

This produces `<timestamp>_<description>.up.sql` and `.down.sql` files. Write your SQL in the up file.

Note: the current down migration is empty. If you add new migrations, write proper down migrations for rollback support.

## SQL Queries and sqlc

SQL queries live in `internal/repository/queries/*.sql`. The service uses [sqlc](https://sqlc.dev/) to generate type-safe Go code from these queries.

When you modify a `.sql` file in `queries/`:

```bash
make sqlc-generate
```

This regenerates Go files in `internal/repository/`:
- `asset.sql.go`
- `ledger.sql.go`
- `transaction.sql.go`
- `user.sql.go`
- `wallet.sql.go`
- `models.go`
- `querier.go`
- `db.go`

**Do not edit these generated files by hand.** They will be overwritten on the next `sqlc generate`.

The only hand-written file in `internal/repository/` is `repository.go`, which provides the `WithTransaction` wrapper.

### sqlc Configuration

From `sqlc.yml`:

```yaml
engine: postgresql
queries: internal/repository/queries/
schema: ./migrations/
sql_package: pgx/v5
```

sqlc reads the migration files to understand the schema, so the migration `.up.sql` files serve as the schema source of truth.

## Adding a New Endpoint

1. Write the SQL query in `internal/repository/queries/<table>.sql`
2. Run `make sqlc-generate`
3. Add service method in `internal/service/`
4. Add handler method in `internal/handler/`
5. Register the route in `internal/router/`
6. Add input validation in `internal/validations/` if needed

## Testing

**There are no tests in this project.** No unit tests, integration tests, or end-to-end tests exist.

TODO: requires implementation. Key areas that would benefit from testing:
- Service layer: topup/spend business logic, idempotency, balance checks
- Repository layer: integration tests against a test database
- Middleware: JWT validation, edge cases (expired, malformed)
- Handler: request/response serialization, validation error formatting

## Debugging

The service uses `zerolog` for structured JSON logging to stderr:

```json
{"level":"info","time":"2026-02-21T08:00:00Z","message":"Server started on http://localhost:8080"}
```

To increase visibility during development, the seed script uses the standard `log` package (not zerolog) and writes to stdout with emoji prefixes.

There are a few `fmt.Println` debug statements in `service/wallet.go` (`TopUp` method) that print balance values. These appear to be development leftovers.
