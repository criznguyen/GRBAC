# Slice 0/1 (MVP) — Completion Summary

**Version:** 1.0  
**Date:** 2025-03-15  
**Status:** Complete (Tech Lead review passed)

---

## 1. What Was Built

### Auth & Middleware (Senior Dev A)
- **Auth middleware** (`internal/middleware/auth.go`): Bearer JWT extraction, HS256/RS256 validation via `JWT_SECRET` or `JWKS_URL`, claims `sub` and `tenant` injected into context
- **Tenant middleware** (`internal/middleware/tenant.go`): X-Tenant-ID validation, tenant existence check, `tenant_id` injected into context
- Unit tests in `internal/middleware/auth_test.go`, `tenant_test.go`

### Tenants & Roles API (Senior Dev B)
- **Tenants** (`internal/api/admin/tenants.go`): POST /tenants, GET /tenants/:id
- **Roles** (`internal/api/admin/roles.go`): POST, GET (list + by id), PUT, DELETE
- **Services** (`internal/service/tenant.go`, `role.go`): Business logic, sqlc-backed
- Unit tests in `internal/api/admin/tenants_test.go`, `roles_test.go`

### Permissions & Audit (Senior Dev C)
- **Permissions** (`internal/api/admin/permissions.go`): PUT/PATCH/DELETE/GET /roles/:id/permissions
- **Audit writer** (`internal/audit/writer.go`): Sync write to `audit_admin` on role/permission mutations
- **Service** (`internal/service/permission.go`): Permission format `resource:action`, upsert via sqlc
- Unit tests in `internal/api/admin/permissions_test.go`, `internal/audit/writer_test.go`, `internal/service/permission_test.go`

### Database
- **Migrations** (`internal/db/migrations/`): Policy Store (tenants, roles, permissions, role_permissions), Audit Store (audit_admin)
- **sqlc** queries in `internal/db/queries/`, generated code in `internal/db/*.sql.go`

---

## 2. How to Run

### Prerequisites
- Go 1.22+
- PostgreSQL 15+
- `DATABASE_URL` env var (e.g. `postgres://user:pass@localhost:5432/grbac?sslmode=disable`)
- `JWT_SECRET` (HS256) or `JWKS_URL` (RS256) for auth

### Steps

```bash
# 1. Install dependencies
go mod tidy

# 2. Run migrations
export DATABASE_URL="postgres://..."
make migrate-up

# 3. Build
make build

# 4. Run the API
export JWT_SECRET="your-secret"  # or JWKS_URL for RS256
make run
```

Server listens on `:8080` (configurable via `SERVER_ADDR`).

### Smoke Test

```bash
# Generate a test JWT (HS256) with payload: {"sub":"test-actor","exp":<future>}
# Use https://jwt.io or your IdP
export SMOKE_TEST_TOKEN="eyJ..."
./scripts/smoke-test.sh
```

Or manually:

```bash
# Health
curl http://localhost:8080/health

# Create tenant (requires Bearer token)
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Acme Corp"}'

# Create role (requires X-Tenant-ID)
curl -X POST http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: <tenant-uuid>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Admin","description":"Full access"}'
```

---

## 3. What's Next

| Slice | Description | Deps |
|-------|-------------|-----|
| 3 | Users, user_roles, groups, group_members, group_roles | Slice 0/1 |
| 4 | Role hierarchy (parent role) | Slice 0/1 |
| 5 | PDP: Check Permission API | Slice 0/1 |
| 6 | Redis cache | Slice 5 |
| 8+ | Tenant bootstrap (built-in roles), delegation, resource registry | Per plan |

See `IMPLEMENTATION-PLAN.md` §9 for full roadmap.

---

## 4. References

- Handoff: `DEV-HANDOFF-TO-SENIOR-DEVS.md`
- Review report: `DEV-REVIEW-REPORT.md`
- API spec: `docs/sdlc/ba/technical/api-spec-admin.md`
