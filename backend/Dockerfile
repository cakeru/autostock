FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

# The migrate CLI is independent of the app source, so install it before
# copying the code in — Docker caches this layer and updates stop recompiling it.
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o autostock ./cmd/server
# One-time storage backfill tool (local disk → object bucket). Run on demand:
#   docker compose run --rm --entrypoint /app/migrate-storage backend [-dry-run]
RUN CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o migrate-storage ./cmd/migrate-storage

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata wget postgresql16-client

COPY --from=builder /app/autostock .
COPY --from=builder /app/migrate-storage .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 -O /dev/null http://localhost:8080/health || exit 1

CMD migrate -path ./migrations -database "$DATABASE_URL" up && ./autostock
