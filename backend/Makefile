# VentureZ backend — common tasks
# Note: on Windows run these via WSL/Git Bash, or use the equivalent commands directly.

MIGRATIONS_DIR := database/migrations
DATABASE_URL ?= postgres://venturez:venturez@localhost:5432/venturez?sslmode=disable

.PHONY: proto up down test lint migrate-up migrate-down migrate-status migrate-reset seed db-setup

## Generate Go + Python gRPC stubs from proto/ into gen/
proto:
	cd proto && buf generate

## Lint and check proto contracts for breaking changes
proto-lint:
	cd proto && buf lint && buf breaking --against '.git#branch=main'

## Spin up all services + postgres + jaeger
up:
	docker compose up --build

down:
	docker compose down -v

## --- Database migrations (goose) ---
## Install once: go install github.com/pressly/goose/v3/cmd/goose@latest

## Apply all pending migrations
migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

## Roll back the most recent migration
migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

## Show migration status
migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

## Roll everything back (destructive)
migrate-reset:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" reset

## Create a new timestamped migration: make migrate-create name=add_widgets
migrate-create:
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

## Load dev seed data (run after migrate-up)
seed:
	psql "$(DATABASE_URL)" -f database/seeds/dev_seed.sql

## Migrate + seed in one shot
db-setup: migrate-up seed

## Run tests across all services
test:
	cd services/gateway && go test ./...
	cd services/agent && python -m pytest
	cd services/financial-engine && go test ./...
