# ADR-001: Technology Choices

## Status

Accepted.

## Context

The system needs to be a wallet microservice that handles financial transactions with correctness and concurrency guarantees. The technology choices need to support:

- Transactional integrity for money movement
- Concurrent request handling
- JSON REST API
- Containerized deployment

## Decision

| Component | Choice | Alternatives considered |
|---|---|---|
| Language | Go 1.25 | Node.js, Java, Rust |
| Database | PostgreSQL | MySQL, SQLite, MongoDB |
| DB driver | pgx/v5 via pgxpool | database/sql + lib/pq |
| Query layer | sqlc (code generation) | GORM, raw SQL strings, sqlx |
| HTTP router | chi/v5 | gin, gorilla/mux, net/http |
| Auth | JWT (golang-jwt, HS256) | Session-based, OAuth |
| Config | koanf (.env + env vars) | viper, envconfig |
| Logger | zerolog | zap, logrus, slog |
| Migrations | golang-migrate | goose, atlas |

## Rationale

**Go**: compiles to a single static binary, has built-in concurrency primitives, and a strong standard library for HTTP. Low runtime overhead compared to JVM or Node.

**PostgreSQL**: ACID compliance, `SELECT ... FOR UPDATE` row-level locking, and mature transaction support. Required for the financial correctness guarantees this system needs. A document database (MongoDB) was ruled out because the data is inherently relational (users → wallets → ledgers).

**pgx over database/sql**: pgx is a native PostgreSQL driver that avoids the overhead of the `database/sql` abstraction. `pgxpool` provides built-in connection pooling with PostgreSQL-specific features (prepared statements, binary protocol).

**sqlc over ORM**: sqlc generates Go code from SQL. This means:
- SQL is visible and auditable (no hidden query generation)
- Compile-time type safety (no runtime SQL construction)
- No N+1 query problems
- Developer must know SQL (a feature, not a drawback, for a financial system)

**chi over gin**: chi is minimalist and fully compatible with `net/http`. It adds routing and middleware without introducing custom context types or response wrappers. gin would also work; this is a preference call.

**zerolog**: zero-allocation structured logging. Other options (zap, slog) are also reasonable. zerolog was chosen for its minimal API.

## Consequences

- Developers must know SQL to modify queries (no ORM query builder).
- Adding a new query requires running `sqlc generate` after modifying `.sql` files.
- pgx ties the service to PostgreSQL specifically (no database portability).
- The Go binary includes no runtime, so Docker images are small (~20MB).
