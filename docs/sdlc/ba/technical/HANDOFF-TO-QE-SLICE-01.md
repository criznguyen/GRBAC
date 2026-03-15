# Handoff to QE — Slice 0/1 Acceptance Criteria

**From:** Dev Teams  
**To:** QE  
**Date:** 2025-03-15  
**Scope:** Slice 0/1 (Policy Store schema + Admin API core)  
**Status:** Ready for Testing

---

## 1. Scope

Slice 0/1 delivers:

- Policy Store DB schema (tenants, roles, permissions, role_permissions)
- Audit Store DB schema (audit_admin)
- Admin API: Tenants, Roles, Permissions (CRUD, role-permission assignment)
- Auth middleware (Bearer token, X-Tenant-ID)
- Admin audit events for role/permission changes

---

## 2. Acceptance Criteria QE Must Test

### 2.1 Tenants

| AC | Description | Pass Criteria |
|----|-------------|---------------|
| T1 | POST /tenants creates tenant | 201; body includes id, name, created_at |
| T2 | POST /tenants with default_policy | Tenant created; deny-all or configurable template applied |
| T3 | GET /tenants/:id returns tenant | 200 with tenant data |
| T4 | GET /tenants/:id for non-existent | 404 |
| T5 | Missing X-Tenant-ID | 401 |
| T6 | Invalid X-Tenant-ID | 401 |

### 2.2 Roles

| AC | Description | Pass Criteria |
|----|-------------|---------------|
| R1 | POST /roles creates role | 201; unique name per tenant |
| R2 | Duplicate role name (same tenant) | 400 |
| R3 | GET /roles returns paginated list | 200; items, next_cursor, has_more |
| R4 | GET /roles/:id returns role | 200 with role data |
| R5 | GET /roles/:id for non-existent | 404 |
| R6 | PUT /roles/:id updates name, description, status | 200 |
| R7 | DELETE /roles/:id removes role | 204; role_permissions cascade-deleted |
| R8 | Tenant isolation: roles from other tenant not returned | Verified |

### 2.3 Permissions (Role-Permission Assignment)

| AC | Description | Pass Criteria |
|----|-------------|---------------|
| P1 | PUT /roles/:id/permissions replaces all | 200; format resource:action or wildcard |
| P2 | PATCH /roles/:id/permissions adds | 200; existing preserved |
| P3 | DELETE /roles/:id/permissions removes | 200 |
| P4 | GET /roles/:id/permissions returns list | 200; array of permission strings |
| P5 | Non-existent role | 404 |
| P6 | Wildcard support (e.g. order:*, *:read) | Accepted and evaluated per spec |

### 2.4 Tenant Isolation

| AC | Description | Pass Criteria |
|----|-------------|---------------|
| I1 | All operations filtered by X-Tenant-ID | No cross-tenant data returned |
| I2 | List endpoints return only tenant-scoped data | Verified |
| I3 | Cross-tenant ID in path (e.g. wrong tenant's role id) | 404 or 403 |

### 2.5 Admin Audit

| AC | Description | Pass Criteria |
|----|-------------|---------------|
| A1 | Create role → audit_admin event | event_type: admin, action_type: create_role |
| A2 | Update role → audit_admin event | action_type: update_role |
| A3 | Delete role → audit_admin event | action_type: delete_role |
| A4 | Assign permissions → audit_admin event | action_type reflects change |
| A5 | Audit includes actor_id, tenant_id, timestamp | All fields present |

---

## 3. Out of Scope (Slice 0/1)

- Users, groups, user_roles, group_roles, group_members
- Role hierarchy (parent/child)
- Check Permission API (PDP)
- Redis cache
- Audit query/export API
- Delegation, resource registry, policy import

---

## 4. Test Case Mapping

| AC | Test Case |
|----|-----------|
| T1–T6 | TC-010 (partial), TC-006 (tenant isolation) |
| R1–R8 | TC-010, TC-011, TC-012 |
| P1–P6 | TC-013 |
| I1–I3 | TC-006 |
| A1–A5 | TC-019 (extend for admin audit) |

---

## 5. References

- **Implementation Plan:** `docs/sdlc/IMPLEMENTATION-PLAN.md`
- **Test Plan:** `docs/sdlc/qe/test-plan-RBAC.md`
- **Test Cases:** `docs/sdlc/qe/test-cases-RBAC.md`
