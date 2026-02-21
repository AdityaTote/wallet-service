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
    -o ./bin/seed \
    ./cmd/seed

FROM alpine:3.22

RUN apk upgrade --no-cache && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder  /build/bin/seed .

COPY --from=builder /build/migrations ./migrations

RUN chown -R appuser:appuser /app

USER appuser

# Run the application
CMD ["./seed"]