# Handoff to Senior Devs — RBAC Slice 0/1

**Version:** 1.0  
**Date:** 2025-03-15  
**From:** Tech Lead  
**Status:** Ready for Implementation

---

## 1. Overview

This document assigns work packages for Slice 0/1 (MVP): Policy Store + Admin API core. Project is bootstrapped; migrations are in place. Senior devs implement handlers, middleware, and audit writer.

---

## 2. Assigned Work Packages

| WP | Description | Owner | Deps |
|----|-------------|-------|------|
| **(1)** | Auth middleware + tenant validation | Senior Dev A | — |
| **(2)** | Tenants API + Roles API | Senior Dev B | WP 1, migrations |
| **(3)** | Permissions API + Audit writer | Senior Dev C | WP 1, WP 2, migrations |

---

## 3. Work Package Details

### 3.1 WP (1): Auth Middleware + Tenant Validation

**Location:** `internal/middleware/`

**Deliverables:**
- `auth.go` — Bearer token extraction, JWT validation (golang-jwt/jwt v5), claims extraction (`sub`, `tenant`)
- `tenant.go` — X-Tenant-ID validation; ensure tenant exists in DB; inject tenant_id into context
- Unit tests for middleware

**Config:** `JWT_SECRET` (HS256) or `JWKS_URL` (RS256) — support both; `JWT_CLAIMS_ISSUER` optional

**API contract:** 401 if missing/invalid token or tenant; 403 if tenant invalid

**Reference:** api-spec-admin.md §2

---

### 3.2 WP (2): Tenants API + Roles API

**Location:** `internal/api/admin/` (or `internal/api/` subdirs)

**Deliverables:**
- `tenants.go` — POST /tenants, GET /tenants/:id
- `roles.go` — POST /roles, GET /roles, GET /roles/:id, PUT /roles/:id, DELETE /roles/:id
- `internal/service/tenant.go`, `role.go` — business logic
- sqlc queries in `internal/db/queries/` (create `sqlc.yaml` and queries)
- Unit tests for handlers

**Endpoints (base `/api/v1`):**

| Method | Path | Handler |
|--------|------|---------|
| POST | /tenants | CreateTenant |
| GET | /tenants/:id | GetTenant |
| POST | /roles | CreateRole |
| GET | /roles | ListRoles (paginated) |
| GET | /roles/:id | GetRole |
| PUT | /roles/:id | UpdateRole |
| DELETE | /roles/:id | DeleteRole |

**Reference:** api-spec-admin.md §3.1, §3.2; IMPLEMENTATION-PLAN.md §6

---

### 3.3 WP (3): Permissions API + Audit Writer

**Location:** `internal/api/admin/`, `internal/audit/`

**Deliverables:**
- `permissions.go` — PUT/PATCH/DELETE/GET /roles/:id/permissions
- `internal/audit/writer.go` — sync write to audit_admin on role/permission changes
- sqlc queries for permissions, role_permissions, audit_admin
- Unit tests

**Endpoints:**

| Method | Path | Handler |
|--------|------|---------|
| PUT | /roles/:id/permissions | ReplaceRolePermissions |
| PATCH | /roles/:id/permissions | AddRolePermissions |
| DELETE | /roles/:id/permissions | RemoveRolePermissions |
| GET | /roles/:id/permissions | ListRolePermissions |

**Permission format:** `resource:action` (e.g. `order:read`, `order:*`). Resolve to permissions table (upsert if not exists).

**Audit events:** Log `action_type` (e.g. `role.create`, `role.update`, `role.delete`, `role.permissions.replace`), `actor_id`, `target_type`, `target_id`, `change_summary`.

**Reference:** api-spec-admin.md §3.3; db-schema.md §3.2; IMPLEMENTATION-PLAN.md §7

---

## 4. File Locations

```
cmd/api/main.go              # Wire routes, middleware, handlers
internal/
├── config/config.go         # Done
├── db/
│   ├── migrations/          # Done (000001, 000002)
│   └── queries/             # sqlc SQL + generated code (you create)
├── middleware/
│   ├── auth.go              # WP 1
│   ├── tenant.go            # WP 1
│   └── middleware.go        # Placeholder
├── api/
│   └── admin/               # WP 2, 3
│       ├── tenants.go
│       ├── roles.go
│       └── permissions.go
├── service/                 # WP 2, 3
│   ├── tenant.go
│   ├── role.go
│   └── permission.go
└── audit/
    ├── writer.go            # WP 3
    └── audit.go             # Placeholder
```

---

## 5. API Contracts Reference

- **Full spec:** `docs/sdlc/ba/technical/api-spec-admin.md`
- **Error format:** §4 — `{"error":{"code":"...","message":"...","request_id":"..."}}`
- **Pagination:** §4 — `limit`, `cursor`; response `items`, `next_cursor`, `has_more`
- **Auth headers:** §2 — `Authorization: Bearer <token>`, `X-Tenant-ID`

---

## 6. DB Schema Reference

- **Policy Store:** `docs/sdlc/ba/technical/db-schema.md` §§2.1, 2.4, 2.5, 2.6
- **Audit Store:** §§3.2
- **Migrations:** `internal/db/migrations/`

---

## 7. Tech Stack Reference

- **Framework:** Chi — `docs/sdlc/DEV-TECH-LEAD-DECISIONS.md`
- **Coding standards:** `docs/sdlc/DEV-CODING-STANDARDS.md`
- **Config:** `internal/config/config.go` — env vars

---

## 8. Review Checklist (Tech Lead Before Merge)

- [ ] `make lint` passes (golangci-lint)
- [ ] `make test` passes
- [ ] Handlers return correct HTTP status codes (201, 200, 404, 400, 401, 403)
- [ ] Tenant isolation: all queries filter by `tenant_id`
- [ ] Error responses match api-spec-admin.md format
- [ ] Audit writer invoked on role/permission mutations
- [ ] No hardcoded secrets; config from env
- [ ] Exported types/functions have doc comments
- [ ] sqlc queries used; no raw SQL in handlers (use generated code)

---

## 9. Getting Started

1. `go mod tidy`
2. `make migrate-up` (set `DATABASE_URL`)
3. Create `sqlc.yaml` (see [sqlc.dev](https://sqlc.dev)) and queries
4. `make sqlc-generate`
5. Implement WP in order: (1) → (2) → (3)

---

## 10. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | Tech Lead | Initial handoff |
