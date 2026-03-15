#!/usr/bin/env bash
# Smoke test for GRBAC Admin API — Slice 0/1
# Requires: server running (make run), DATABASE_URL, JWT_SECRET, SMOKE_TEST_TOKEN
#
# Generate a test JWT (HS256) with sub + tenant claims, e.g.:
#   Header: {"alg":"HS256","typ":"JWT"}
#   Payload: {"sub":"test-actor","tenant":"<tenant-uuid>","exp":9999999999}
# Use https://jwt.io or: echo -n '{"sub":"test-actor","exp":9999999999}' | base64
#
# Usage: SMOKE_TEST_TOKEN=<jwt> BASE_URL=${BASE_URL:-http://localhost:8080} ./scripts/smoke-test.sh

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN="${SMOKE_TEST_TOKEN:?Set SMOKE_TEST_TOKEN to a valid JWT}"

echo "=== Smoke Test: GRBAC Admin API ==="
echo "Base URL: $BASE_URL"
echo ""

# 1. Health check
echo "[1/6] GET /health"
curl -sf "$BASE_URL/health" | jq .
echo ""

# 2. Create tenant
echo "[2/6] POST /api/v1/tenants"
TENANT_RESP=$(curl -sf -X POST "$BASE_URL/api/v1/tenants" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Smoke Test Tenant"}')
TENANT_ID=$(echo "$TENANT_RESP" | jq -r '.id')
echo "$TENANT_RESP" | jq .
echo "Tenant ID: $TENANT_ID"
echo ""

# 3. Create role (requires X-Tenant-ID)
echo "[3/6] POST /api/v1/roles"
ROLE_RESP=$(curl -sf -X POST "$BASE_URL/api/v1/roles" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{"name":"Editor","description":"Can edit resources"}')
ROLE_ID=$(echo "$ROLE_RESP" | jq -r '.id')
echo "$ROLE_RESP" | jq .
echo "Role ID: $ROLE_ID"
echo ""

# 4. Assign permissions
echo "[4/6] PUT /api/v1/roles/$ROLE_ID/permissions"
PERM_RESP=$(curl -sf -X PUT "$BASE_URL/api/v1/roles/$ROLE_ID/permissions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{"permissions":["order:read","order:write","product:*"]}')
echo "$PERM_RESP" | jq .
echo ""

# 5. Get role permissions
echo "[5/6] GET /api/v1/roles/$ROLE_ID/permissions"
curl -sf "$BASE_URL/api/v1/roles/$ROLE_ID/permissions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" | jq .
echo ""

# 6. Get tenant
echo "[6/6] GET /api/v1/tenants/$TENANT_ID"
curl -sf "$BASE_URL/api/v1/tenants/$TENANT_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" | jq .
echo ""

echo "=== Smoke test PASSED ==="
