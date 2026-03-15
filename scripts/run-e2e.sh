#!/usr/bin/env bash
# E2E test: start Postgres (Docker), migrate, start API, run smoke-test, cleanup.
# Requires: Docker, migrate, jq, curl. Optional: make, go (for gen-jwt).
#
# Usage: ./scripts/run-e2e.sh
# Env overrides: E2E_DB_PORT (default 5433), BASE_URL (default http://localhost:8080)

set -e

E2E_DB_PORT="${E2E_DB_PORT:-5433}"
BASE_URL="${BASE_URL:-http://localhost:8080}"
CONTAINER_NAME="${E2E_DB_NAME:-grbac-e2e-db}"
JWT_SECRET="${JWT_SECRET:-test-secret}"
DATABASE_URL="postgres://postgres:postgres@localhost:${E2E_DB_PORT}/grbac_test?sslmode=disable"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
API_PID=""

cleanup() {
  if [[ -n "$API_PID" ]] && kill -0 "$API_PID" 2>/dev/null; then
    echo "Stopping API (PID $API_PID)..."
    kill "$API_PID" 2>/dev/null || true
    wait "$API_PID" 2>/dev/null || true
  fi
  if docker ps -q -f "name=^${CONTAINER_NAME}$" 2>/dev/null | grep -q .; then
    echo "Stopping Postgres container..."
    docker stop "$CONTAINER_NAME" 2>/dev/null || true
    docker rm -f "$CONTAINER_NAME" 2>/dev/null || true
  fi
}
trap cleanup EXIT

echo "=== E2E: GRBAC Admin API ==="

# Prefer binary, else go run
gen_jwt() {
  if [[ -x "$ROOT_DIR/bin/gen-jwt" ]]; then
    JWT_SECRET="$JWT_SECRET" SUB="${SUB:-test-actor}" TENANT_ID="${TENANT_ID:-}" EXP_HOURS="${EXP_HOURS:-1}" "$ROOT_DIR/bin/gen-jwt"
  else
    (cd "$ROOT_DIR" && JWT_SECRET="$JWT_SECRET" SUB="${SUB:-test-actor}" TENANT_ID="${TENANT_ID:-}" EXP_HOURS="${EXP_HOURS:-1}" go run ./cmd/gen-jwt)
  fi
}

command -v docker >/dev/null 2>&1 || { echo "docker required"; exit 1; }
command -v migrate >/dev/null 2>&1 || { echo "migrate required (make install-tools or install golang-migrate)"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "jq required"; exit 1; }
command -v curl >/dev/null 2>&1 || true || { echo "curl required"; exit 1; }

# Start Postgres
if docker ps -q -f "name=^${CONTAINER_NAME}$" 2>/dev/null | grep -q .; then
  echo "Using existing container $CONTAINER_NAME"
else
  echo "Starting Postgres on port $E2E_DB_PORT..."
  docker run -d --name "$CONTAINER_NAME" \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_DB=grbac_test \
    -p "${E2E_DB_PORT}:5432" \
    postgres:16-alpine
  sleep 3
fi

# Migrate
echo "Running migrations..."
(cd "$ROOT_DIR" && DATABASE_URL="$DATABASE_URL" make migrate-up)

# Build API and gen-jwt
echo "Building API..."
(cd "$ROOT_DIR" && make build)
if ! [[ -x "$ROOT_DIR/bin/gen-jwt" ]]; then
  echo "Building gen-jwt..."
  (cd "$ROOT_DIR" && go build -o bin/gen-jwt ./cmd/gen-jwt)
fi

# Start API in background
export DATABASE_URL
export AUDIT_DATABASE_URL="$DATABASE_URL"
export JWT_SECRET
export SERVER_ADDR=":8080"
echo "Starting API..."
"$ROOT_DIR/bin/grbac-api" &
API_PID=$!

# Wait for health
echo "Waiting for API health..."
for i in $(seq 1 30); do
  if curl -sf "$BASE_URL/health" >/dev/null 2>&1; then
    echo "API is up."
    break
  fi
  if ! kill -0 "$API_PID" 2>/dev/null; then
    echo "API process exited unexpectedly"
    exit 1
  fi
  sleep 1
done
curl -sf "$BASE_URL/health" >/dev/null || { echo "API health check failed"; exit 1; }

# Generate JWT and run smoke test
export SMOKE_TEST_TOKEN
SMOKE_TEST_TOKEN="$(gen_jwt)"
export BASE_URL
echo "Running smoke test..."
"$SCRIPT_DIR/smoke-test.sh"

echo "=== E2E PASSED ==="
