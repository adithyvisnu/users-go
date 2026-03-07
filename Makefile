.PHONY: run build docs lint test docker-up docker-down

## Install swaggo CLI (run once)
install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest

## Generate Swagger docs from code annotations → docs/
docs:
	swag init -g cmd/api/main.go -o docs
	@echo "✅ Swagger docs regenerated — visit http://localhost:8080/swagger/index.html"

## Run locally (generates docs first)
run: docs
	go run ./cmd/api

## Build binary
build: docs
	go build -ldflags="-w -s" -o bin/users-api ./cmd/api

## Run tests
test:
	go test ./... -v -race -cover

## Lint
lint:
	golangci-lint run ./...

## Start all services via Docker Compose
docker-up:
	docker compose up --build

## Stop all services
docker-down:
	docker compose down -v
