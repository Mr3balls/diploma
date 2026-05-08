APP_NAME=esports-api
DB_URL=postgres://postgres:postgres@localhost:5432/esports?sslmode=disable

.PHONY: run build test fmt migrate-up migrate-down docker-up docker-down seed

run:
	go run ./cmd/api

build:
	go build -o bin/$(APP_NAME) ./cmd/api

test:
	go test ./...

fmt:
	gofmt -w $(shell find . -name '*.go')

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

seed:
	psql "$(DB_URL)" -f migrations/000002_seed.sql

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v
