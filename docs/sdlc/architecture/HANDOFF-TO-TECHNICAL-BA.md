# Handoff to Technical BA — RBAC for Microservices

**From:** Architect  
**To:** Technical BA  
**Date:** 2025-03-15  
**Input docs:** HANDOFF-TO-ARCHITECT.md, FRS-RBAC.md, process-flows-RBAC.md, glossary-RBAC.md, and all architecture artifacts in `docs/sdlc/architecture/`

---

## 1. Purpose of This Handoff

This document summarizes the architecture decisions and deliverables produced by the Architect. It tells the Technical BA what is fixed by architecture, what remains to be specified in detail, and what constraints to respect when producing API specs, DB schema, and team breakdown.

---

## 2. Architecture Deliverables (Produced)

| Artifact | Location | Content |
|----------|----------|---------|
| System Context (C4 L1) | `system-context.md` | Actors, RBAC system boundary, IdP and Microservices (PEP) as external systems. |
| Container Diagram (C4 L2) | `container-diagram.md` | Admin API, PDP, Policy Store, Audit Store, Cache; data flows. |
| ADR-001 | `ADR-001-PDP-topology.md` | Hybrid: central PDP + SDK optional local cache; no sidecar in Phase 1. |
| ADR-002 | `ADR-002-data-store.md` | Policy Store and Audit Store both PostgreSQL; audit in separate DB. |
| ADR-003 | `ADR-003-caching-strategy.md` | Redis for PDP; decision cache; key shape; TTL 60s default; invalidation by subject/role/tenant. |
| ADR-004 | `ADR-004-audit-storage-retention.md` | Append-only audit in separate PostgreSQL; async write for checks; retention configurable (min 1 year); partitioning. |
| ADR-005 | `ADR-005-multi-tenancy.md` | Row-level isolation with tenant_id; no schema-per-tenant in Phase 1; scope as optional filter. |
| Tech stack | `tech-stack.md` | PostgreSQL, Redis, REST/gRPC, optional message queue; language/framework left to Technical BA or team. |

---

## 3. Key Architecture Decisions (Constraints for Technical BA)

- **PDP:** Centralized PDP; microservices call it via REST or gRPC. SDKs are thin clients; optional in-process cache with TTL. No sidecar PDP in Phase 1.
- **Data stores:** Two PostgreSQL databases (or two DBs in one cluster): Policy Store (roles, permissions, assignments, groups, delegation, registry) and Audit Store (append-only, partitioned by time). All policy and audit tables include tenant_id.
- **Cache:** Redis shared by all PDP instances; cache permission check results; key includes tenant, subject, resource, action, scope; default TTL 60s; Admin API triggers invalidation on every relevant policy/assignment change.
- **Audit:** Every permission check and every admin action is logged. Check audit write is async (off critical path) so p99 < 50ms is achievable; admin audit can be sync or async. No update/delete on audit rows; retention and partitioning as in ADR-004.
- **Multi-tenancy:** Single schema with tenant_id; every API and query is tenant-scoped; scope (org/project) is optional on assignments and check requests.

---

## 4. What Technical BA Must Specify

### 4.1 APIs

- **Admin API (REST):** Full specification of endpoints, request/response bodies, and status codes for:
  - Roles: create, update, delete, list (FR-001–FR-003).
  - Permissions: assign to role (FR-004).
  - Assignments: assign role to user, assign role to group (FR-005, FR-006); optional scope.
  - Role hierarchy: set parent role (FR-007).
  - Groups: CRUD, add/remove members (FR-008, FR-009).
  - Delegation: create/update/delete delegation rules (FR-010).
  - Resource registry: register resources/actions (FR-019).
  - Policy as code: import from YAML/JSON (FR-020).
  - Built-in roles and default policy: tenant bootstrap (FR-021, FR-022).
  - Audit: query and export (FR-018); filters, pagination, format (CSV/JSON).
- **Check Permission API:** Request/response contract (subject, resource, action, tenant, optional scope → allowed: boolean); batch check format; REST and/or gRPC.
- **Auth:** How tenant and subject are supplied (headers, JWT claims); validation and error responses (401/403).

### 4.2 Data Model and DB Schema

- **Policy Store:** Normalized schema for tenants, roles, permissions, role_permission, user_role, group_role, groups, group_membership, delegation, role_hierarchy, resource_registry. All with tenant_id; indexes (tenant_id, …) for list and lookup. Technical BA to define exact columns, types, and constraints.
- **Audit Store:** Schema for check events and admin events; partition key (e.g. month); indexes for (tenant_id, event_time), (subject_id, event_time), (resource, action), (actor_id) to support query and export.
- **Retention and partitioning:** Exact partition strategy and retention job behavior (e.g. drop partitions older than N months).

### 4.3 Cache and Invalidation

- **Redis key namespace and key format** (e.g. `rbac:check:{tenant_id}:{subject_id}:{resource}:{action}:{scope_hash}`).
- **Invalidation:** How Admin API triggers invalidation (e.g. Redis DEL by pattern, or version key per subject); which events trigger which invalidation (role update, assign/revoke user/group, group membership change, etc.).
- **TTL:** Configurable; default 60s; document in API or config spec.

### 4.4 Async Audit Write

- **Mechanism:** In-process queue + background writer, or message queue (e.g. Redis Streams, Kafka). At-least-once delivery; backpressure and failure handling.
- **Audit event payload:** Exact fields for check events and admin events (tenant_id, subject_id, resource, action, decision, timestamp, request_id, scope; actor_id, action_type, target, change_summary, etc.).

### 4.5 SDK Contracts

- **Node, Go, Java:** Method signature (e.g. `check(userId, resource, action, options?)`); how tenant and scope are passed; timeout and retry; optional local cache (TTL, key shape). Document alignment with Check Permission API.

### 4.6 Team Breakdown

- Use or adapt template: `docs/sdlc/ba/technical/team-breakdown.template.md`.
- Break work into implementable slices (e.g. Policy Store + Admin API core CRUD; PDP + Check API + Redis cache; Audit Store + async writer + query/export; SDKs; delegation and scope; resource registry and policy import). Dependencies (e.g. schema first, then API, then PDP) should be clear.

---

## 5. NFR Checklist for Technical BA

When specifying APIs and schema, ensure:

- **p99 < 50ms (Check Permission):** Cache hit path and async audit write; avoid blocking on audit DB in check response.
- **99.9% availability:** Stateless PDP and Admin API; DB and Redis HA; idempotency and timeouts/retries in SDK.
- **Audit compliance:** 100% of checks and admin actions logged; immutable audit store; retention and export as specified in ADR-004.
- **Multi-tenancy:** No cross-tenant data; tenant_id mandatory and validated on every request.

---

## 6. References

| Document | Path | Use |
|----------|------|-----|
| BA handoff | `docs/sdlc/ba/business/HANDOFF-TO-ARCHITECT.md` | Requirements and constraints |
| FRS | `docs/sdlc/ba/business/FRS-RBAC.md` | All 24 FRs |
| Process flows | `docs/sdlc/ba/business/process-flows-RBAC.md` | Assign Role, Check Permission, Audit Query/Export |
| Glossary | `docs/sdlc/ba/business/glossary-RBAC.md` | Terms |
| System context | `docs/sdlc/architecture/system-context.md` | C4 L1 |
| Container diagram | `docs/sdlc/architecture/container-diagram.md` | C4 L2 |
| ADRs 001–005 | `docs/sdlc/architecture/ADR-001-*.md` … `ADR-005-*.md` | Decisions |
| Tech stack | `docs/sdlc/architecture/tech-stack.md` | Technologies |

---

## 7. Handoff Checklist

- [x] System context and container diagram produced
- [x] ADRs 001–005 for PDP, data store, cache, audit, multi-tenancy
- [x] Tech stack documented
- [x] Constraints and “what to specify” clearly listed for Technical BA
- [x] NFR alignment and references provided

**Next phase:** Technical BA produces API specifications, DB schema, cache/invalidation and audit-write details, SDK contracts, and team breakdown in `docs/sdlc/ba/technical/`.
