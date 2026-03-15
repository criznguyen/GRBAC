# RBAC Test Plan

**Version:** 1.0  
**Date:** 2025-03-15  
**Author:** QE  
**Input docs:** FRS-RBAC.md, HANDOFF-TO-DEV.md, api-spec-admin.md, api-spec-check-permission.md, api-spec-audit.md  
**Status:** Ready for Execution

---

## 1. Scope

This test plan covers all functional requirements (FR-001–FR-024) for the RBAC for Microservices product. Test areas are mapped to FRS epics and traceability is maintained FR → Test area → Test case IDs.

---

## 2. FRS to Test Area Mapping

| FR | Epic | Test Area | Test Type | Test Case IDs |
|----|------|-----------|-----------|---------------|
| FR-001 | Epic 1 | Create Role | Integration, E2E | TC-003, TC-010 |
| FR-002 | Epic 1 | Update Role | Integration, E2E | TC-011 |
| FR-003 | Epic 1 | Delete Role | Integration, E2E | TC-012 |
| FR-004 | Epic 1 | Assign Permissions to Role | Integration, E2E | TC-003, TC-013 |
| FR-005 | Epic 1 | Assign Role to User | Integration, E2E | TC-003 |
| FR-006 | Epic 1 | Assign Role to Group | Integration, E2E | TC-014 |
| FR-007 | Epic 1 | Role Inheritance | Integration, E2E | TC-004 |
| FR-008 | Epic 2 | Create/Manage Groups | Integration, E2E | TC-015 |
| FR-009 | Epic 2 | Add/Remove Users from Group | Integration, E2E | TC-016 |
| FR-010 | Epic 2 | Delegation | Integration, E2E, Security | TC-017 |
| FR-011 | Epic 3 | Check Permission API | Integration, E2E, Performance | TC-001, TC-002, TC-007, TC-008 |
| FR-012 | Epic 3 | SDK for Permission Check | Integration | TC-018 |
| FR-013 | Epic 3 | PDP Standalone | Integration, E2E | TC-007 |
| FR-014 | Epic 4 | Multi-Tenancy | Integration, E2E, Security | TC-006 |
| FR-015 | Epic 4 | Scope for Roles | Integration, E2E | TC-009 |
| FR-016 | Epic 5 | Audit Log for Permission Checks | Integration, E2E | TC-005 |
| FR-017 | Epic 5 | Audit Log for Admin Actions | Integration, E2E | TC-019 |
| FR-018 | Epic 5 | Export Audit Log | Integration, E2E | TC-020 |
| FR-019 | Epic 6 | Register Resources and Actions | Integration | TC-021 |
| FR-020 | Epic 6 | Policy as Code (Import) | Integration, E2E | TC-022 |
| FR-021 | Epic 7 | Built-in Roles | Integration | TC-023 |
| FR-022 | Epic 7 | Default Policy for New Tenant | Integration | TC-024 |
| FR-023 | Epic 8 | Caching for PDP | Integration, Performance | TC-007 |
| FR-024 | Epic 8 | Horizontal Scaling | Performance, E2E | TC-025 |

---

## 3. Test Types

| Type | Purpose | Tools / Approach |
|------|---------|------------------|
| **Unit** | Services, PDP evaluation logic, permission matching | Jest, Vitest, or language-native |
| **Integration** | API endpoints, DB, Redis, audit writer | Supertest, test containers |
| **E2E** | Full flows: create tenant → role → assign → check | Playwright, Cypress, or API E2E |
| **Performance** | p99 < 50ms check path, throughput | k6, Artillery, wrk |
| **Security** | Tenant isolation, auth, delegation boundaries | OWASP-style, penetration tests |

---

## 4. Test Areas Detail

### 4.1 Admin API (FR-001–FR-010, FR-019–FR-022)

| Area | Scope | Key scenarios |
|------|-------|---------------|
| Tenants | CRUD, bootstrap | Create with default policy, get by id |
| Roles | CRUD, hierarchy | Create, update, delete; set parent; no circular ref |
| Permissions | Assign to role | PUT replace, PATCH add, DELETE remove, GET list |
| Users | CRUD, assign roles | Create, list, assign/revoke roles with scope |
| Groups | CRUD, members, roles | Create, add/remove members, assign/revoke roles |
| Delegations | CRUD, enforce | Create rule; delegated user assign/revoke in scope |
| Resource Registry | CRUD | Register resource types and actions |
| Policy Import | Import, dry-run | YAML/JSON import; dry_run preview |

### 4.2 Check Permission (FR-011, FR-013, FR-023)

| Area | Scope | Key scenarios |
|------|-------|---------------|
| Single check | POST /check | Allow, Deny, scope filter |
| Batch check | POST /check/batch | Order preserved; independent results |
| Evaluation | Logic | User roles + group roles + inheritance; deny by default |
| Cache | Hit/miss | Cache hit returns within p99; invalidation on change |

### 4.3 Multi-Tenancy & Scope (FR-014, FR-015)

| Area | Scope | Key scenarios |
|------|-------|---------------|
| Tenant isolation | All APIs | No cross-tenant data; X-Tenant-ID validated |
| Scope | Assignments, check | Scope filter on check; org/project scoped roles |

### 4.4 Audit (FR-016, FR-017, FR-018)

| Area | Scope | Key scenarios |
|------|-------|---------------|
| Check audit | 100% of checks | Allow and Deny logged; queryable |
| Admin audit | All admin changes | Actor, action type, target, change summary |
| Export | CSV/JSON | Filters, pagination; async for large exports |

### 4.5 Performance & Scalability (FR-023, FR-024)

| Area | Scope | Key scenarios |
|------|-------|---------------|
| Latency | Check path | p99 < 50ms on cache hit |
| Throughput | Load | Linear scaling with instances |
| Availability | HA | 99.9% under load |

---

## 5. Environment Requirements

- Policy Store (PostgreSQL)
- Audit Store (PostgreSQL, separate)
- Redis (for PDP cache)
- Test tenant(s) and admin user(s)
- IdP / JWT for auth (or test token generator)

---

## 6. Traceability Matrix (Summary)

| FR | Epic | Test Case IDs |
|----|------|---------------|
| FR-001 | Core RBAC | TC-003, TC-010 |
| FR-002 | Core RBAC | TC-011 |
| FR-003 | Core RBAC | TC-012 |
| FR-004 | Core RBAC | TC-003, TC-013 |
| FR-005 | Core RBAC | TC-003 |
| FR-006 | Groups | TC-014 |
| FR-007 | Core RBAC | TC-004 |
| FR-008 | Groups | TC-015 |
| FR-009 | Groups | TC-016 |
| FR-010 | Delegation | TC-017 |
| FR-011 | Policy Decision | TC-001, TC-002, TC-007, TC-008 |
| FR-012 | Policy Decision | TC-018 |
| FR-013 | Policy Decision | TC-007 |
| FR-014 | Multi-Tenancy | TC-006 |
| FR-015 | Multi-Tenancy | TC-009 |
| FR-016 | Audit | TC-005 |
| FR-017 | Audit | TC-019 |
| FR-018 | Audit | TC-020 |
| FR-019 | Resource Registry | TC-021 |
| FR-020 | Policy as Code | TC-022 |
| FR-021 | Built-in Roles | TC-023 |
| FR-022 | Default Policy | TC-024 |
| FR-023 | Performance | TC-007 |
| FR-024 | Scalability | TC-025 |

---

## 7. References

- **FRS:** `docs/sdlc/ba/business/FRS-RBAC.md`
- **Handoff:** `docs/sdlc/ba/technical/HANDOFF-TO-DEV.md`
- **Test Cases:** `docs/sdlc/qe/test-cases-RBAC.md`
- **Deploy Checklist:** `docs/sdlc/qe/HANDOFF-TO-DEPLOY.md`

---

## 8. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | QE | Initial test plan |
