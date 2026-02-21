# Operations & Observability

## Health Checks

### Application Health Endpoint

`GET /api/health` returns service status and a database connectivity check:

```json
{
  "status": "healthy",
  "timestamp": "2026-02-21T08:00:00Z",
  "checks": {
    "database": {
      "status": "healthy",
      "response_time": "1.2ms"
    }
  }
}
```

The check pings the PostgreSQL connection pool. If the ping fails, `status` is `"unhealthy"`.

### Docker Health Check

The `wallet.dockerfile` includes:

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1
```

Docker marks the container unhealthy if the health endpoint fails 3 consecutive checks.

## Logging

The wallet service uses `zerolog` for structured JSON logging to stderr:

```json
{"level":"info","time":"...","message":"Server started on http://localhost:8080"}
{"level":"error","error":"...","message":"failed to top-up wallet"}
```

Log levels used in the codebase:
- `Info` — startup, health check results
- `Debug` — idempotent retry detection, auth rejections
- `Error` — failed operations (topup, spend, balance retrieval, auth)
- `Fatal` — startup failures (config, database connection)

The seed script (`cmd/seed/main.go`) uses the standard Go `log` package, not zerolog.

### What Is Not Logged

- Request/response bodies
- Request IDs or correlation IDs
- Latency per request
- Client IP addresses

TODO: structured request logging middleware (method, path, status, latency) is not implemented.

## Metrics

No metrics are instrumented. There is no Prometheus endpoint, no StatsD, no OpenTelemetry.

TODO: if metrics are needed, the following would be useful:
- Request count and latency by endpoint
- Transaction success/failure rate
- Database connection pool utilization (`pgxpool` exposes stats via `pool.Stat()`)
- Active database transactions

## Scaling

### Horizontal Scaling

The service is stateless. Multiple instances can run behind a load balancer, all pointing to the same PostgreSQL. Concurrency safety is handled at the database level (`SELECT ... FOR UPDATE`), so this works correctly across instances.

### Connection Pooling

The `pgxpool` defaults are used. The current codebase does not configure pool size explicitly. The pgx defaults are:
- `MaxConns`: 4 (or `max(4, numCPU)` depending on version)
- `MinConns`: 0

TODO: for production load, you likely need to tune `MaxConns` based on expected concurrency and PostgreSQL `max_connections`.

### Database Scaling

PostgreSQL is the bottleneck for write-heavy workloads. The `FOR UPDATE` lock serializes mutations per wallet, so throughput scales linearly with the number of distinct wallets being accessed concurrently. Two users transacting on their own wallets never contend.

## Backups

No backup strategy is configured. The Docker Compose volume (`pg_vol`) persists data locally but is not backed up.

TODO: requires a backup strategy (pg_dump cron job, WAL archiving, or managed PostgreSQL).

## Graceful Shutdown

The application listens for `SIGINT`:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
// ...
<-ctx.Done()
```

However, examining `main()`, after receiving the signal, it calls `stop()` and `cancel()` but does **not** call `srv.Shutdown(ctx)`. This means in-flight HTTP requests may be terminated abruptly. The `Shutdown` method exists in `server.go` but is never invoked.

TODO: wire up `srv.Shutdown(ctx)` in `main()` for proper graceful shutdown.
