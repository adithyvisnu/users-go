.PHONY: run build docs lint test docker-up docker-down

# PODMAN_COMPOSE := python3 $(HOME)/Library/Python/3.9/bin/Podman-compose.py

PODMAN_COMPOSE := python3 $(HOME)/Library/Python/3.9/lib/python/site-packages/podman_compose.py

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

## Start all services via Podman
docker-up:
	$(PODMAN_COMPOSE) up --build

## Stop all services
docker-down:
	$(PODMAN_COMPOSE) down -v
