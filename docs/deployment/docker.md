# Docker Deployment

## Images

Two Dockerfiles in `docker/`, both using multi-stage builds:

### wallet.dockerfile

- **Builder stage**: `golang:1.25.7-alpine` — downloads deps, compiles a static binary (`CGO_ENABLED=0`)
- **Runtime stage**: `alpine:3.22` — copies binary and migrations, runs as non-root user (`appuser:1000`)
- Includes a `HEALTHCHECK` that hits `GET /api/health` every 30s
- Entrypoint: `./wallet-service`

### seed.dockerfile

- Same build pattern as above
- No health check (it's a run-once job)
- Entrypoint: `./seed`

Both produce static Linux/amd64 binaries with no runtime Go dependency.

## Docker Compose

`docker-compose.yml` defines four services:

| Service | Image / Build | Role |
|---|---|---|
| `postgres` | `postgres:latest` | Database. Hardcoded creds: `postgres`/`postgres`/`postgres` |
| `migrate` | `migrate/migrate` | Runs `UP` migrations. Depends on postgres (healthy) |
| `seed-service` | Built from `docker/seed.dockerfile` | Seeds data. Depends on migrate |
| `wallet-service` | Built from `docker/wallet.dockerfile` | API server on port 8080. Depends on seed-service |

### Startup Order

```
postgres (health check passes)
  └─▶ migrate (runs migrations, exits)
       └─▶ seed-service (seeds data, exits)
            └─▶ wallet-service (runs indefinitely)
```

`migrate` has `restart: on-failure` so it retries if postgres isn't fully ready despite the health check.

### Volumes

- `pg_vol` — persists PostgreSQL data at `/var/lib/postgresql`

### Port Mappings

| Host | Container | Service |
|---|---|---|
| 5432 | 5432 | postgres |
| 8080 | 8080 | wallet-service |

### Environment

All services read from `.env` via `env_file: .env`. The postgres service additionally sets its own credentials via `environment:`.

## Commands

```bash
# Start everything
docker compose up --build

# Start in background
docker compose up --build -d

# View logs
docker compose logs -f wallet-service

# Stop and remove containers (keep data)
docker compose down

# Stop, remove containers, AND delete database volume
docker compose down -v

# Rebuild a single service
docker compose build wallet-service
```

## Production Considerations

The current Docker Compose setup is designed for local development. For production:

- **Database credentials** are hardcoded in `docker-compose.yml`. Use secrets management.
- **PostgreSQL image** is `postgres:latest`. Pin to a specific version.
- **No TLS**. The HTTP server has no TLS configuration.
- **No resource limits** on containers.
- **No log rotation** configuration.
- **No external volume backup** strategy.
- **No container registry**. Images are built locally.

TODO: production deployment (Kubernetes, ECS, etc.) is not configured.
