# Integration Tests — RBAC Admin API

Integration tests for the RBAC Admin API that hit a real API server and real PostgreSQL database. Uses **testcontainers-go** to spin up an ephemeral Postgres container.

## Requirements

- **Docker** — must be running (testcontainers spawns a Postgres container)
- **Go 1.25+**

## Running

From the project root:

```bash
make test-integration
```

Or directly:

```bash
go test -race -count=1 -tags=integration ./tests/integration/...
```

## Scope

Integration tests are excluded from `go test ./...` and `make test` by the `//go:build integration` tag, so unit tests remain fast and do not require Docker.

## What's Tested

| Test | Coverage |
|------|----------|
| `TestIntegration_Tenants_CreateAndGet` | Create tenant (201), get by ID (200), response shape |
| `TestIntegration_Roles_CreateListGetUpdateDelete` | Full CRUD: create, list, get, update, delete role |
| `TestIntegration_Permissions_AssignAndList` | Replace, add, remove permissions; list role permissions |
| `TestIntegration_TenantIsolation` | Create 2 tenants; role in T1 not visible to T2 |
| `TestIntegration_Audit_AdminEvents` | Create/update role, assign permissions → verify `audit_admin` entries |

## Structure

- **`setup.go`** — `SetupTestDB(t)` spins up Postgres via testcontainers, runs migrations, returns `(dbURL, cleanup)`
- **`server.go`** — `StartTestServer(t, dbURL)` builds Chi router with real services, returns `httptest.Server`
- **`jwt.go`** — `NewTestJWT(sub, tenantID, secret)` for HS256 Bearer tokens
- **`client.go`** — `APIClient` for API calls with automatic `Authorization` and `X-Tenant-ID`
- **`audit.go`** — `QueryAuditAdminEvents(t, dbURL, tenantID)` to assert audit entries
- **`admin_api_test.go`** — integration test cases

## Troubleshooting

- **Docker not running** — Start Docker Desktop (or equivalent) before `make test-integration`
- **Connection refused** — Ensure no firewall blocks Docker networking
- **Migrations path** — Tests must be run from the project root (default for `go test`)
