# Handoff to Architect — RBAC for Microservices

**From:** Business BA  
**To:** Architect  
**Date:** 2025-03-15  
**Input docs:** PRD-RBAC-Microservices.md, epic-brief-RBAC.md, FRS-RBAC.md, process-flows-RBAC.md, glossary-RBAC.md

---

## 1. Purpose of This Handoff

This document summarizes what the Architect needs to know to produce architecture decisions, ADRs, and technical design for the RBAC system. It highlights key functional requirements, constraints, non-functional requirements, and process flows that will drive system design.

---

## 2. Scope Summary

- **What we're building:** A central RBAC service for microservices: role and permission model, policy decision point (PDP), audit, multi-tenancy, SDKs, optional resource registry and policy-as-code.
- **Out of scope (Phase 1):** Authentication (integrate with existing IdP); full ABAC/ReBAC; admin UI (possible Phase 2).

---

## 3. Key Functional Requirements (FRS) — Design Drivers

| Area | Key FRs | Design implication |
|------|---------|---------------------|
| **Core RBAC** | FR-001–FR-007: CRUD roles, assign permissions to roles, assign roles to user/group, role inheritance | Data model: Role, Permission, User, Group, UserRole, GroupRole, RolePermission, RoleHierarchy. Resolution: expand roles with inheritance and group membership. |
| **Groups & delegation** | FR-008–FR-010: Groups, membership, delegation rules | Store delegation rules (delegator, role, scope); enforce scope when delegator performs assign/revoke. |
| **Policy decision** | FR-011–FR-013: Check Permission API, SDK, PDP standalone | Stateless PDP; clear separation control plane vs data plane; API contract: subject, resource, action → allowed. |
| **Multi-tenancy & scope** | FR-014–FR-015: Tenant isolation, optional scope (org/project) | Every entity and API keyed by tenant_id; scope on assignments and on check requests. |
| **Audit** | FR-016–FR-018: Log every check + every admin action; query and export | Append-only audit store; query/export API; retention (e.g. 1 year) and compliance (GDPR, PCI-DSS, SOC2). |
| **Resource registry** | FR-019–FR-020: Register resources/actions; policy as code import | Registry store; optional validation of permission strings; import pipeline for YAML/JSON. |
| **Built-in & default** | FR-021–FR-022: Built-in roles per tenant; default policy (deny-all/template) on tenant creation | Bootstrap logic on tenant create; built-in roles identifiable and protected. |
| **Performance** | FR-023–FR-024: Caching (TTL + invalidation on change); horizontal scaling | Cache layer (edge/sidecar or central); invalidation on role/permission/assignment change; stateless PDP behind LB. |

---

## 4. Process Flows to Support

1. **Assign Role to User/Group** — API that creates user-role or group-role assignments; validation (tenant, scope, delegation); audit; cache invalidation. See `process-flows-RBAC.md` §1.
2. **Check Permission (Allow/Deny)** — Resolve subject → user + groups → roles (with inheritance) → permissions; evaluate (resource, action); optional cache; audit every check. See `process-flows-RBAC.md` §2.
3. **Audit Query / Export** — Query by filters (date, user, resource, action, tenant); export CSV/JSON with pagination; authorization for Auditor role. See `process-flows-RBAC.md` §3.

These flows should be reflected in component interaction and data flow (e.g. PDP → cache → audit write).

---

## 5. Non-Functional Requirements (from PRD)

| NFR | Requirement |
|-----|-------------|
| **Latency** | p99 < 50ms for permission check end-to-end |
| **Availability** | 99.9% SLA for RBAC service |
| **Security** | Zero-trust; least privilege; support JWT/OAuth2 token validation (subject/tenant from token) |
| **Data** | Audit log retained at least 1 year (configurable) |
| **Compliance** | Support GDPR, PCI-DSS, SOC2 audit needs (immutable audit, query/export) |

---

## 6. Constraints and Assumptions

- **Dependencies:** IdP supplies user identity (OAuth2/OIDC); RBAC receives subject id (and optionally tenant) from token. Optional: message queue for cache invalidation.
- **Integration:** Microservices call RBAC via HTTP/gRPC (sync) or via sidecar; tenant id from context (header, token claim, or routing).
- **Consistency:** Cache invalidation must happen on every relevant policy/assignment change so that permission results stay correct (FR-023).
- **Stateless PDP:** No per-request state on PDP instances; shared store and cache for policy and decisions.

---

## 7. Key Design Decisions to Make (Architect)

- **Data store:** Choice of DB for roles, permissions, assignments, groups, delegation, resource registry; and separate or same store for audit (append-only).
- **PDP topology:** Co-located cache vs central cache; sidecar vs in-process SDK; gRPC vs REST for check API.
- **Cache strategy:** Key shape (tenant+subject+resource+action+scope); TTL; invalidation (per-subject, per-role, per-tenant, or broadcast).
- **Audit storage:** Schema, retention, and indexing for query/export; sync vs async write; impact on p99.
- **Multi-tenancy:** Single DB with tenant_id vs tenant-scoped schemas/databases; impact on isolation and scaling.
- **Delegation:** How scope (e.g. group_id, org_id) is stored and evaluated when a delegated user performs assign/revoke.

---

## 8. Deliverables Expected from Architect

- Architecture overview and component diagram (control plane, data plane, audit, cache).
- ADRs for major choices (store, cache, PDP deployment, audit backend).
- Data model (logical and key entities) and API boundaries.
- Deployment and scaling approach (stateless PDP, LB, horizontal scaling).
- Security and compliance considerations (zero-trust, token validation, audit immutability).

---

## 9. Document References

| Document | Location | Use |
|----------|----------|-----|
| PRD | `docs/sdlc/po/PRD-RBAC-Microservices.md` | Epics, user stories, NFRs, glossary |
| Epic brief | `docs/sdlc/po/epic-brief-RBAC.md` | Priorities, success metrics, risks |
| FRS | `docs/sdlc/ba/business/FRS-RBAC.md` | All 24 FRs with traceability |
| Process flows | `docs/sdlc/ba/business/process-flows-RBAC.md` | Assign Role, Check Permission, Audit Query/Export |
| Glossary | `docs/sdlc/ba/business/glossary-RBAC.md` | Terms and definitions |

---

## 10. Handoff Checklist

- [x] FRS covers all 8 Epics and traces to Epic/US
- [x] Process flows for Assign Role, Check Permission, Audit Query/Export documented
- [x] Glossary expanded from PRD
- [x] NFRs and constraints summarized for Architect
- [x] Expected Architect deliverables listed

**Next phase:** Architect produces architecture artifacts in `docs/sdlc/architecture/`.
