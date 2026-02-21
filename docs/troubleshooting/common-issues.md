# Troubleshooting

## Service Won't Start

### "failed to load config"

The application panics with `panic("failed to load config")`.

**Cause**: missing required environment variables. All of `PORT`, `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_NAME`, `JWT_SECRET` must be set.

**Fix**: ensure `.env` exists and contains all variables from `.env.example`. In Docker, ensure `env_file: .env` is configured.

### "failed to initiate db"

The application logs a fatal error and exits.

**Cause**: cannot connect to PostgreSQL, or the connection ping timed out (3 second timeout in `database.go`).

**Fix**:
- Verify PostgreSQL is running and reachable at the configured host:port
- Verify credentials match
- In Docker Compose, the `migrate` service depends on `postgres` health check. If you're running locally, ensure PostgreSQL is up before starting the service

### Port already in use

**Cause**: another process is using port 8080.

**Fix**: change `PORT` in `.env`, or stop the conflicting process.

## Migration Issues

### "dirty database version"

golang-migrate marks the database as dirty if a migration fails partway through.

**Fix**:
```bash
migrate -path migrations -database "postgresql://..." force <version>
```

Replace `<version>` with the last successfully applied version number (or 0 to reset).

### "migrate" container keeps restarting

The `restart: on-failure` policy causes the migrate container to retry if it fails. This can happen if PostgreSQL is slow to become healthy.

**Fix**: usually resolves itself once PostgreSQL passes the health check. If stuck, check `docker compose logs migrate` for the actual error.

## Seed Issues

### "Database already seeded! Skipping..."

The seed script detected existing data and did not modify anything.

**Fix**: if you need to re-seed, delete existing data first:
```sql
DELETE FROM ledgers;
DELETE FROM transactions;
DELETE FROM wallets;
DELETE FROM users;
DELETE FROM assets;
```
Then run `make seed` or restart the Docker Compose stack.

## Authentication Issues

### "unauthorized" on wallet endpoints

**Possible causes**:
1. Missing `Authorization` header
2. Header format is not `Bearer <token>` (must be exactly one space, case-sensitive `Bearer`)
3. Token is expired (24h TTL)
4. Token was signed with a different `JWT_SECRET` (e.g. secret changed since token was issued)
5. User or wallet was deleted from the database after token was issued

**Debug**: decode the JWT at [jwt.io](https://jwt.io) to inspect claims and expiration.

### "authentication failed" on signup

**Cause**: username already exists. The error message is intentionally generic to avoid leaking user enumeration information.

## Transaction Issues

### "transaction failed" on topup/spend

This is a generic error wrapping the actual database failure. Check the service logs (stderr) for specifics.

**Common causes**:
- Database connection pool exhausted
- PostgreSQL connection dropped
- The system wallet does not exist (seed was not run)

### Idempotent retry returns unexpected balance

If you retry with the same `txn_id` and get a different balance than expected, this is correct behavior. The idempotent retry returns the **current** balance at the time of the retry, which may differ from the balance at the time of the original transaction if other transactions occurred in between.

### Spend fails despite seemingly sufficient balance

**Possible cause**: concurrent transactions. Between reading the balance in your client and submitting the spend, another spend may have reduced the balance below the required amount. The `FOR UPDATE` lock ensures the balance check within the transaction is accurate, but a client-side balance check is inherently stale.

## Docker Issues

### "wallet-service" container exits immediately

Check logs: `docker compose logs wallet-service`. Common causes:
- Seed service hasn't completed yet
- Database connection failed (check `DATABASE_HOST` — inside Docker it should be `postgres`, not `localhost`)

### Cannot connect to the API from the host

The service binds to `:8080` inside the container. The port mapping `"8080:8080"` exposes it on the host. If you cannot connect:
- Verify the container is running: `docker compose ps`
- Verify the health check is passing: `docker inspect --format='{{.State.Health.Status}}' wallet-service`
- Check if another service is using port 8080 on the host

## Database Issues

### `database.Close()` recursive call

`internal/database/database.go` has a bug in the `Close()` method:

```go
func (db *Database) Close() error {
    db.Close()  // recursive call — this will stack overflow
    return nil
}
```

This method calls itself instead of `db.Pool.Close()`. It is only called during shutdown (`server.Shutdown`), which is itself never called from `main()`. If graceful shutdown were wired up, this would cause a stack overflow.

**Fix**: change `db.Close()` to `db.Pool.Close()` in `database.go`.
