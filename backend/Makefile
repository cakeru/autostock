.PHONY: run test migrate-up migrate-down sqlc build tidy

run:
	go run ./cmd/server

test:
	go test ./... -v

test-cover:
	go test ./... -cover

migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

migrate-reset:
	migrate -path ./migrations -database "$(DATABASE_URL)" reset

sqlc:
	sqlc generate

build:
	go build -o bin/autostock ./cmd/server

tidy:
	go mod tidy
	go mod verify

lint:
	golangci-lint run

dev:
	air -- --port=8080
