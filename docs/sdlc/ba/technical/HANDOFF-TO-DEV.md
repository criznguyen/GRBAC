# Handoff to Dev Teams — RBAC for Microservices

**From:** Technical BA  
**To:** Dev Teams (Backend, SDK, DevOps)  
**Date:** 2025-03-15  
**Input docs:** Architect handoff, FRS, ADRs, API specs, DB schema, team breakdown

---

## 1. Purpose

This document summarizes the technical specifications for implementation. Dev teams use it as the single handoff for API contracts, DB schema, implementation order, and acceptance criteria.

---

## 2. Input Documents

| Document | Path | Content |
|----------|------|---------|
| Admin API Spec | `api-spec-admin.md` | All Admin API endpoints, request/response schemas |
| Check Permission API Spec | `api-spec-check-permission.md` | POST /check, batch, gRPC optional |
| Audit API Spec | `api-spec-audit.md` | Query, export; filters, pagination |
| DB Schema | `db-schema.md` | Policy Store + Audit Store DDL, indexes, partitioning |
| Team Breakdown | `team-breakdown.md` | Slices, dependencies, acceptance criteria |
| Architecture | `docs/sdlc/architecture/` | Container diagram, ADRs, tech stack |
| FRS | `docs/sdlc/ba/business/FRS-RBAC.md` | Functional requirements |

---

## 3. API Contracts Summary

### 3.1 Admin API (Base: `/api/v1`)

| Area | Key Endpoints | Auth |
|------|---------------|------|
| Tenants | POST /tenants, GET /tenants/{id} | Bearer + X-Tenant-ID |
| Roles | POST/GET/PUT/DELETE /roles, PUT /roles/{id}/parent | Same |
| Permissions | PUT/PATCH/DELETE/GET /roles/{id}/permissions | Same |
| Users | POST/GET /users, POST/DELETE /users/{id}/roles | Same |
| Groups | CRUD /groups, POST/DELETE /groups/{id}/members, POST/DELETE /groups/{id}/roles | Same |
| Delegations | CRUD /delegations | Same |
| Resource Registry | POST/GET/PUT /resources | Same |
| Policy Import | POST /policy/import | Same |

**Headers:** `Authorization: Bearer <token>`, `X-Tenant-ID: <tenant_id>`

### 3.2 Check Permission API

| Endpoint | Method | Request | Response |
|----------|--------|---------|----------|
| /check | POST | `{ subject_id, resource, action, scope? }` | `{ allowed: boolean }` |
| /check/batch | POST | `{ checks: [...] }` | `{ results: [{ allowed }...] }` |

**Headers:** Same as Admin API.

### 3.3 Audit API (Base: `/api/v1/audit`)

| Endpoint | Method | Purpose |
|----------|--------|---------|
| /audit/events | GET | Query with filters, pagination |
| /audit/export | POST | Export CSV/JSON; sync or async job |

---

## 4. DB Schema Summary

### Policy Store (PostgreSQL)

| Table | Purpose |
|-------|---------|
| tenants | Tenant registry |
| users | User references (tenant-scoped) |
| groups | Groups |
| roles | Roles (with is_builtin) |
| permissions | Canonical resource:action |
| role_permissions | Role ↔ permission |
| role_hierarchy | Parent-child roles |
| user_roles | User ↔ role (with scope) |
| group_roles | Group ↔ role (with scope) |
| group_members | Group ↔ user |
| delegations | Delegation rules |
| resource_registry | Resource types + actions |

### Audit Store (PostgreSQL, separate DB)

| Table | Purpose |
|-------|---------|
| audit_checks | Permission check events (partitioned by event_time) |
| audit_admin | Admin action events (partitioned by event_time) |

**Indexes:** See `db-schema.md` for full list. Key: tenant_id, subject/actor lookups, date range for audit.

---

## 5. Implementation Order

| Phase | Deliverable | Owner |
|-------|-------------|-------|
| 1 | DB schema migration (Policy + Audit) | Backend |
| 2 | Admin API core (tenants, roles, permissions) | Backend |
| 3 | Admin API (users, groups, assignments) | Backend |
| 4 | PDP + Check Permission API | Backend |
| 5 | Redis cache + invalidation | Backend |
| 6 | Delegation, resource registry, policy import | Backend |
| 7 | Audit async writer + Audit API | Backend |
| 8 | SDKs (Node, Go, Java) | SDK |
| 9 | Infra: PostgreSQL, Redis, deployment, retention | DevOps |

Details and dependencies: `team-breakdown.md`.

---

## 6. Acceptance Criteria by Component

### 6.1 Admin API

- [ ] All CRUD endpoints return correct HTTP status codes (201, 200, 204, 400, 401, 403, 404)
- [ ] Tenant isolation: no cross-tenant data; X-Tenant-ID validated
- [ ] Role hierarchy: no circular reference; inheritance evaluated correctly
- [ ] Assign/revoke triggers cache invalidation
- [ ] Policy import supports dry-run; atomic or phased apply

### 6.2 Check Permission API (PDP)

- [ ] Single check returns `{ allowed: true|false }` per evaluation logic
- [ ] Batch check returns results in same order as input
- [ ] Deny by default when no matching permission
- [ ] Scope filter applied when provided
- [ ] p99 < 50ms on cache hit path
- [ ] 100% of checks logged to audit (async)

### 6.3 Audit API

- [ ] Query returns paginated results with filters (date, subject, resource, actor, event_type)
- [ ] Export returns CSV or JSON; async for large exports
- [ ] Authorization: only Auditor/Admin can query/export

### 6.4 Cache

- [ ] Redis key: `rbac:check:{tenant_id}:{subject_id}:{resource}:{action}:{scope_hash}`
- [ ] TTL configurable; default 60s
- [ ] Invalidation on every policy/assignment change (subject, role, tenant)

### 6.5 SDKs

- [ ] `check(userId, resource, action, options?)` aligns with API
- [ ] Tenant and scope passed via options
- [ ] Timeout and retry configurable
- [ ] Integration achievable in < 1 day per service (per PRD)

### 6.6 Infrastructure

- [ ] PostgreSQL HA (Policy + Audit stores)
- [ ] Redis HA (cluster or Sentinel)
- [ ] Stateless Admin API + PDP; horizontal scaling
- [ ] Audit retention job: drop partitions older than configurable period (min 1 year)

---

## 7. NFR Checklist

| NFR | Implementation |
|-----|----------------|
| p99 < 50ms (Check) | Cache hit path; async audit write |
| 99.9% availability | Stateless APIs; DB/Redis HA; LB |
| Audit compliance | 100% checks + admin logged; immutable; retention |
| Multi-tenancy | tenant_id on all tables; validated on every request |

---

## 8. References

- **Architecture:** `docs/sdlc/architecture/HANDOFF-TO-TECHNICAL-BA.md`, ADRs 001–005
- **FRS:** `docs/sdlc/ba/business/FRS-RBAC.md`
- **Process flows:** `docs/sdlc/ba/business/process-flows-RBAC.md`

---

## 9. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | Technical BA | Initial handoff to Dev |
