# RBAC for Microservices — Makefile
# Tech Lead decisions: Chi, sqlc, golang-migrate, pgx
#
# Test targets:
#   test             - unit tests (no Docker)
#   test-integration - integration tests (Docker + testcontainers)
#   test-all         - test + test-integration
#   test-e2e         - full E2E: DB, migrate, API, smoke-test (Docker, migrate, jq)

BINARY_NAME ?= grbac-api
MAIN_PATH   := ./cmd/api
MIGRATIONS  := ./internal/db/migrations

.PHONY: build build-gen-jwt test test-integration test-all test-e2e run lint migrate-up migrate-down migrate-create tidy sqlc-generate

# Build the API binary
build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

# Build JWT generator for scripts/smoke/e2e (optional)
build-gen-jwt:
	go build -o bin/gen-jwt ./cmd/gen-jwt

# Run unit tests (excludes integration; no Docker required)
test:
	go test -race -count=1 ./...

# Run integration tests (requires Docker; uses testcontainers for PostgreSQL)
test-integration:
	go test -race -count=1 -tags=integration ./tests/integration/...

# Run all automation: unit + integration (sequential)
test-all: test test-integration

# E2E: start Postgres (Docker), migrate, start API, run smoke-test, cleanup. Requires Docker, migrate, jq.
test-e2e:
	./scripts/run-e2e.sh

# Run tests with coverage
test-coverage:
	go test -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run the API (requires DATABASE_URL, etc.)
run: build
	./bin/$(BINARY_NAME)

# Run the API without building (for dev)
run-dev:
	go run $(MAIN_PATH)/main.go

# Lint with golangci-lint (uses go run if binary not in PATH)
lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run ./...

# Run migrations up
migrate-up:
	migrate -path $(MIGRATIONS) -database "$(DATABASE_URL)" up

# Run migrations down
migrate-down:
	migrate -path $(MIGRATIONS) -database "$(DATABASE_URL)" down 1

# Create a new migration (usage: make migrate-create name=add_users)
migrate-create:
	@test -n "$(name)" || (echo "Usage: make migrate-create name=<migration_name>" && exit 1)
	migrate create -ext sql -dir $(MIGRATIONS) -seq $(name)

# Generate sqlc code (requires sqlc.yaml)
# Uses go run if sqlc not in PATH
sqlc-generate:
	@command -v sqlc >/dev/null 2>&1 && sqlc generate || go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

# Tidy go.mod
tidy:
	go mod tidy

# Install tools (golangci-lint, migrate, sqlc)
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
