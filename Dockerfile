FROM golang:1.24-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/server ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/bin/server /app/server

# Default env values for local/docker-compose runs.
ENV DATABASE_URL=postgres://root:password@postgres:5432/my_db?sslmode=disable \
    OPERATOR_API_KEY=dev-operator-key \
    REDIS_ADDR=redis:6379 \
    REDIS_PASSWORD="" \
    REDIS_DB=0 \
    WEBHOOK_URL=http://host.docker.internal:9090/webhook \
    STATS_TIME_WINDOW_MINUTES=60 \
    CACHE_INCIDENTS_TTL_SECONDS=60 \
    WEBHOOK_MAX_RETRIES=5 \
    WEBHOOK_RETRY_BASE_SECONDS=10

EXPOSE 8080
ENTRYPOINT ["/app/server"]
