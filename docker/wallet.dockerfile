FROM golang:1.25.7-alpine AS builder

RUN apk add --no-cache --upgrade git ca-certificates tzdata && \
    apk upgrade --no-cache

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a \
    -o ./bin/wallet-service \
    ./cmd/wallet-service

FROM alpine:3.22

RUN apk upgrade --no-cache && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder  /build/bin/wallet-service .

COPY --from=builder /build/migrations ./migrations

RUN chown -R appuser:appuser /app

USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the application
CMD ["./wallet-service"]