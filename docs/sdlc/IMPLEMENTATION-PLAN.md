# RBAC Implementation Plan — MVP (Slice 0/1)

**Version:** 1.0  
**Date:** 2025-03-15  
**From:** Dev Teams  
**Input docs:** HANDOFF-TO-DEV.md, team-breakdown.md, db-schema.md, api-spec-admin.md, tech-stack.md  
**Status:** Ready for Development

---

## 1. Overview

This document defines the implementation plan for the RBAC MVP (Slice 0/1): **Policy Store schema + Admin API core** covering Tenants, Roles, Permissions, and Role-Permission assignment. Subsequent slices build on this foundation.

---

## 2. Scope: Slice 0/1 (MVP)

| Deliverable | Description | FRs |
|-------------|-------------|-----|
| **DB schema** | Policy Store + Audit Store migrations | — |
| **Admin API core** | Tenants, Roles, Permissions, Role-Permission CRUD | FR-001, FR-002, FR-003, FR-004, FR-014, FR-021, FR-022 |
| **Auth middleware** | Bearer token + X-Tenant-ID validation | FR-014 |
| **Audit (admin)** | Admin action events written to Audit Store | FR-017 |

---

## 3. Implementation Order

Following HANDOFF-TO-DEV.md §5 and team-breakdown.md §3.2:

| Phase | Task | Owner | Dependencies |
|-------|------|-------|--------------|
| **1** | DB schema migration (Policy Store) | Backend | — |
| **2** | DB schema migration (Audit Store) | Backend | — |
| **3** | Auth middleware (Bearer, X-Tenant-ID) | Backend | — |
| **4** | Tenants API (POST, GET) | Backend | 1, 3 |
| **5** | Roles API (CRUD, list) | Backend | 1, 3, 4 |
| **6** | Permissions model + role_permissions | Backend | 5 |
| **7** | Role-Permission API (PUT, PATCH, DELETE, GET) | Backend | 6 |
| **8** | Admin audit writer (sync for MVP) | Backend | 2, 3 |
| **9** | Tenant bootstrap (built-in roles, default policy) | Backend | 4, 5, 6, 7 |

---

## 4. Tech Stack (per tech-stack.md)

| Layer | Choice | Notes |
|-------|--------|-------|
| **Runtime** | Node.js 20+ | Alternative: Go 1.22+ per team preference |
| **Framework** | Express or Fastify | REST JSON; middleware for auth |
| **Policy Store** | PostgreSQL 15+ | ACID, tenant_id on all tables |
| **Audit Store** | PostgreSQL (separate DB) | Append-only, partitioned |
| **ORM / Migrations** | node-pg + node-pg-migrate or Prisma | Or: sqlx (Go), Flyway (Java) |
| **Auth** | JWT validation (jsonwebtoken or jose) | Extract sub, tenant from token |
| **Validation** | express-validator or Zod | Request validation |

---

## 5. File Structure (Scaffold)

```
src/
├── api/                 # REST endpoints
│   ├── admin/           # Admin API routes
│   │   ├── tenants.ts
│   │   ├── roles.ts
│   │   ├── permissions.ts
│   │   └── index.ts
│   └── index.ts
├── pdp/                 # Policy Decision Point (future)
│   └── README.md
├── audit/               # Audit writer & events
│   ├── writer.ts
│   └── types.ts
├── db/                  # DB layer
│   ├── policy-store.ts  # Policy Store client
│   ├── audit-store.ts   # Audit Store client
│   └── migrations/      # SQL migrations
│       ├── 001_policy_store.sql
│       └── 002_audit_store.sql
├── sdk/                 # SDK packages (future)
│   └── README.md
├── middleware/          # Auth, tenant, error handling
│   ├── auth.ts
│   └── tenant.ts
├── services/            # Business logic
│   ├── tenant.service.ts
│   ├── role.service.ts
│   └── permission.service.ts
├── config/              # Configuration
│   └── index.ts
├── app.ts               # Express app setup
└── server.ts            # Entry point
```

---

## 6. Slice 0/1 API Endpoints

| Method | Path | Purpose | FR |
|--------|------|---------|-----|
| POST | /api/v1/tenants | Create tenant (with default policy) | FR-021, FR-022 |
| GET | /api/v1/tenants/:id | Get tenant | FR-014 |
| POST | /api/v1/roles | Create role | FR-001 |
| GET | /api/v1/roles | List roles (paginated) | FR-001 |
| GET | /api/v1/roles/:id | Get role | FR-001 |
| PUT | /api/v1/roles/:id | Update role | FR-002 |
| DELETE | /api/v1/roles/:id | Delete role | FR-003 |
| PUT | /api/v1/roles/:id/permissions | Replace role permissions | FR-004 |
| PATCH | /api/v1/roles/:id/permissions | Add role permissions | FR-004 |
| DELETE | /api/v1/roles/:id/permissions | Remove role permissions | FR-004 |
| GET | /api/v1/roles/:id/permissions | List role permissions | FR-004 |

---

## 7. DB Tables (Slice 0/1)

### Policy Store (required for MVP)

- `tenants`
- `roles`
- `permissions`
- `role_permissions`

### Audit Store (required for MVP)

- `audit_admin` (partitioned by event_time)

### Deferred to later slices

- users, groups, role_hierarchy, user_roles, group_roles, group_members, delegations, resource_registry
- audit_checks (when PDP is implemented)

---

## 8. Acceptance Criteria for QE (Slice 0/1)

1. **Tenants**
   - POST /tenants creates tenant with default policy; returns 201 with id, name, created_at.
   - GET /tenants/:id returns tenant or 404.
   - Missing/invalid X-Tenant-ID returns 401.

2. **Roles**
   - POST /roles creates role with unique name per tenant; returns 201.
   - Duplicate role name returns 400.
   - GET /roles returns paginated list; GET /roles/:id returns single role.
   - PUT /roles/:id updates name, description, status; returns 200.
   - DELETE /roles/:id returns 204; role_permissions cascade-deleted.

3. **Permissions**
   - PUT /roles/:id/permissions replaces all permissions; format resource:action or wildcard.
   - PATCH /roles/:id/permissions adds permissions; DELETE removes.
   - GET /roles/:id/permissions returns current permissions.
   - Non-existent role returns 404.

4. **Tenant isolation**
   - All operations filtered by X-Tenant-ID.
   - No cross-tenant data returned or modified.

5. **Admin audit**
   - Create/update/delete role and permission changes logged to audit_admin.

---

## 9. Subsequent Slices (Reference)

| Slice | Description | Deps |
|-------|-------------|------|
| 3 | Users, user_roles, groups, group_members, group_roles | Slice 2 |
| 4 | Role hierarchy (role_hierarchy, set parent API) | Slice 2 |
| 5 | PDP: Check Permission (single + batch) | Slice 2 |
| 6 | Redis cache | Slice 5 |
| 7 | Cache invalidation | Slice 3, 6 |
| 8–13 | Delegation, resource registry, policy import, audit writer/API, tenant bootstrap | Per team-breakdown |

---

## 10. References

- **Technical BA:** `docs/sdlc/ba/technical/HANDOFF-TO-DEV.md`, `api-spec-admin.md`, `db-schema.md`, `team-breakdown.md`
- **Architecture:** `docs/sdlc/architecture/tech-stack.md`, ADRs 001–005
- **FRS:** `docs/sdlc/ba/business/FRS-RBAC.md`

---

## 11. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | Dev Teams | Initial implementation plan for Slice 0/1 |
