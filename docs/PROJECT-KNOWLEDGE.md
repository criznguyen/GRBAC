# RBAC Microservices — Project Knowledge Base

**Mục đích:** Tài liệu tham chiếu tổng hợp cho AI và team. Đọc file này để nắm toàn bộ ngữ cảnh dự án.

**Cập nhật:** 2025-03-15

---

## 1. Tổng quan dự án

| Thuộc tính | Giá trị |
|------------|---------|
| **Tên** | RBAC (Role-Based Access Control) for Microservices |
| **Mục tiêu** | Hệ thống RBAC trung tâm cho microservices, enterprise + SME |
| **Phạm vi** | Roles, Permissions, Users, Groups, PDP, Audit, Multi-tenancy, SDK |
| **Workflow** | SDLC: PO → Business BA → Architect → Technical BA → Dev → QE → Deploy |
| **Trạng thái** | Docs đầy đủ; Implementation Slice 0/1 (MVP) ready |

---

## 2. Cấu trúc tài liệu (docs/sdlc/)

```
po/                    # PRD, Epic brief
ba/business/           # FRS (24 FRs), process flows, glossary
architecture/          # System context, container diagram, 5 ADRs, tech stack
ba/technical/          # API specs (admin, check, audit), db-schema, team-breakdown
qe/                    # Test plan, test cases (TC-001–TC-025), deploy checklist
IMPLEMENTATION-PLAN.md # Slice 0/1 MVP, Go stack, file structure
```

---

## 3. Epics & User Stories (tóm tắt)

| Epic | Mô tả | Priority |
|------|-------|----------|
| E1 | Core RBAC: Roles, Permissions, Users/Groups, Role inheritance | Must |
| E2 | Groups, Delegation | Should |
| E3 | Check Permission API, SDK, PDP standalone | Must |
| E4 | Multi-tenancy, Scope | Must |
| E5 | Audit (check + admin, export) | Must |
| E6 | Resource Registry, Policy as Code | Should |
| E7 | Built-in roles, Default policy | Should |
| E8 | Caching, Horizontal scaling | Must |

---

## 4. Functional Requirements (24 FRs)

- **FR-001–FR-007:** Core RBAC (roles CRUD, permissions, user/group assign, inheritance)
- **FR-008–FR-010:** Groups, delegation
- **FR-011–FR-013:** Check Permission API, SDK, PDP standalone
- **FR-014–FR-015:** Multi-tenancy, scope
- **FR-016–FR-018:** Audit (checks, admin, export)
- **FR-019–FR-020:** Resource registry, policy import
- **FR-021–FR-022:** Built-in roles, default policy
- **FR-023–FR-024:** Caching, horizontal scaling

---

## 5. Kiến trúc

### Containers
- **Admin API:** Control plane (CRUD roles, permissions, assignments, groups, delegation, registry)
- **PDP:** Data plane (Check Permission, Allow/Deny)
- **Policy Store:** PostgreSQL (roles, permissions, users, groups, …)
- **Audit Store:** PostgreSQL (riêng DB, append-only, partitioned)
- **Cache:** Redis (decision cache; SME: in-memory fallback)

### ADRs chính
- **ADR-001:** Hybrid PDP — central PDP, SDK optional local cache
- **ADR-002:** PostgreSQL cho Policy + Audit (2 DB riêng)
- **ADR-003:** Redis cache, key `rbac:check:{tenant}:{subject}:{resource}:{action}:{scope_hash}`, TTL 60s
- **ADR-004:** Audit async write; retention ≥ 1 năm; partitioned
- **ADR-005:** Multi-tenancy = row-level `tenant_id`; no schema-per-tenant Phase 1

---

## 6. Công nghệ

| Thành phần | Chọn |
|------------|------|
| Runtime | **Go 1.22+** |
| Framework | Chi hoặc Echo |
| DB | PostgreSQL (Policy + Audit) |
| Cache | Redis (+ in-memory fallback cho SME) |
| DB layer | sqlc + golang-migrate |
| Deploy SME | Docker Compose |
| Deploy Enterprise | Kubernetes |

---

## 7. API chính

### Admin API (`/api/v1`)
- Tenants: POST/GET
- Roles: CRUD, PUT /roles/{id}/parent, PUT/PATCH/DELETE/GET /roles/{id}/permissions
- Users, Groups, Delegations, Resource Registry, Policy Import

### Check Permission API
- `POST /check` — `{ subject_id, resource, action, scope? }` → `{ allowed }`
- `POST /check/batch` — `{ checks: [...] }` → `{ results: [...] }`

### Audit API
- `GET /audit/events` — Query với filters, pagination
- `POST /audit/export` — CSV/JSON, sync hoặc async job

---

## 8. DB Schema (Policy Store)

- tenants, users, groups, roles, permissions
- role_permissions, role_hierarchy
- user_roles, group_roles, group_members
- delegations, resource_registry

Tất cả có `tenant_id`.

### Audit Store
- audit_checks (partitioned by event_time)
- audit_admin (partitioned)

---

## 9. NFRs

| NFR | Target |
|-----|--------|
| Latency p99 (Check) | < 50ms |
| Availability | 99.9% |
| Audit | 100% logged, retain ≥ 1 năm |
| Time to integrate | < 1 ngày/service |

---

## 10. Implementation Plan (Slice 0/1 MVP)

1. DB schema (Policy + Audit)
2. Auth middleware (Bearer, X-Tenant-ID)
3. Tenants API
4. Roles API (CRUD)
5. Permissions + role_permissions
6. Role-Permission API (PUT, PATCH, DELETE, GET)
7. Admin audit writer
8. Tenant bootstrap

**File structure (Go):**
```
cmd/api/main.go
internal/api/admin/, pdp/
internal/audit/, db/, middleware/, service/, config/
pkg/
```

---

## 11. Test & Deploy

- **Test plan:** FR-001–FR-024 → test areas
- **Test cases:** TC-001–TC-025 (Allow, Deny, Assign Role, Inheritance, Audit, Multi-tenancy, Cache, Batch, Scope, …)
- **Deploy checklist:** Smoke S1–S12, performance baseline, infra checks, rollback plan

---

## 12. Glossary (tóm tắt)

- **Role:** Tập permissions; job function
- **Permission:** (Resource, Action) — e.g. `order:read`
- **Subject:** User/service account cần check
- **PDP:** Policy Decision Point (Allow/Deny)
- **PEP:** Policy Enforcement Point (trong microservice)
- **Tenant:** Đơn vị isolation

---

## 13. Handoff chain

```
PO (PRD, Epic) → Business BA (FRS, flows) → Architect (ADRs, diagrams)
  → Technical BA (API, DB, team) → Dev (Implementation) → QE (Test, Deploy)
```

---

## 14. Chi tiết bổ sung (đã đọc & ghi nhận)

### Process flows chính
- **Assign Role to User:** Validate tenant → authorize caller → validate user/role → create assignment → audit → invalidate cache
- **Check Permission:** Validate request → cache lookup → resolve subject+groups → resolve roles (incl. inheritance) → permissions → evaluate (resource, action) → audit → return `{ allowed }`
- **Audit Query/Export:** Authorize Auditor → validate filters → query Audit Store → paginate / export CSV|JSON

### Slice 0/1 endpoints cụ thể
| Method | Path |
|--------|------|
| POST | /api/v1/tenants |
| GET | /api/v1/tenants/:id |
| POST | /api/v1/roles |
| GET | /api/v1/roles, /api/v1/roles/:id |
| PUT | /api/v1/roles/:id |
| DELETE | /api/v1/roles/:id |
| PUT/PATCH/DELETE/GET | /api/v1/roles/:id/permissions |

### Headers bắt buộc
- `Authorization: Bearer <token>` — JWT, extract `sub`, optional `tenant`
- `X-Tenant-ID` — tenant context (nếu không có trong token)

### Personas
- System Admin, Service Owner, Application Developer, Auditor

### Tech recommendation (enterprise + SME)
- SME: Docker Compose, in-memory cache, ~$0–20/tháng
- Enterprise: K8s, Redis Cluster, RDS Multi-AZ, $200+/tháng
- Cùng codebase, khác config/deployment

### MCP / mcp-brain
- `.cursor/mcp.json` config: DATABASE_URL, MEMORY_NAMESPACE=project/GRBAC
- migrate đã chạy thành công; Cursor cần bật mcp-brain trong MCP settings để dùng memory tools

---

*File này phục vụ recall nhanh. Chi tiết xem các docs tương ứng trong `docs/sdlc/`.*
