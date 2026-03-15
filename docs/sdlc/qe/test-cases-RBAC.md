# RBAC Test Cases — Critical Paths

**Version:** 1.0  
**Date:** 2025-03-15  
**Author:** QE  
**Input docs:** FRS-RBAC.md, api-spec-check-permission.md, api-spec-admin.md  
**Template:** test-case.template.md

---

## TC-001: Check Permission — Allow

**Precondition**: Tenant exists. User has role with `order:read` permission. User is assigned the role (direct or via group).

**Steps**:
1. Create tenant T1 (if not exists)
2. Create role "Order Viewer" with permission `order:read`
3. Create user U1 in tenant T1
4. Assign role "Order Viewer" to user U1
5. Call `POST /api/v1/check` with `{ subject_id: U1, resource: "order", action: "read" }`, headers `Authorization: Bearer <valid_token>`, `X-Tenant-ID: T1`
6. Assert response status 200
7. Assert `response.body.allowed === true`

**Expected**: Response `{ allowed: true }`; status 200.

**Links to**: FR-011, FR-013

---

## TC-002: Check Permission — Deny

**Precondition**: Tenant exists. User has no role with `order:delete` permission, or user has no matching permission.

**Steps**:
1. Create tenant T1 (if not exists)
2. Create user U1 in tenant T1 (no role assignment, or role without `order:delete`)
3. Call `POST /api/v1/check` with `{ subject_id: U1, resource: "order", action: "delete" }`, headers `Authorization: Bearer <valid_token>`, `X-Tenant-ID: T1`
4. Assert response status 200
5. Assert `response.body.allowed === false`

**Expected**: Response `{ allowed: false }`; status 200. Deny by default when no matching permission.

**Links to**: FR-011

---

## TC-003: Assign Role to User

**Precondition**: Tenant exists. User and role exist in same tenant.

**Steps**:
1. Create tenant T1
2. Create role "Order Manager" with permissions `order:read`, `order:create`, `order:update`
3. Create user U1 in tenant T1
4. Call `POST /api/v1/users/U1/roles` with `{ role_ids: ["<Order Manager role id>"] }`, headers `Authorization`, `X-Tenant-ID: T1`
5. Assert response 201 or 200
6. Call `POST /api/v1/check` with subject_id: U1, resource: "order", action: "create"
7. Assert `allowed === true`

**Expected**: Role assigned; subsequent check returns Allow for `order:create`.

**Links to**: FR-005, FR-001, FR-004

---

## TC-004: Role Inheritance

**Precondition**: Tenant exists. Child role has parent role; parent has permissions child does not.

**Steps**:
1. Create tenant T1
2. Create role "Viewer" with `order:read`
3. Create role "Editor" with `order:create`, `order:update`
4. Set "Editor" parent to "Viewer" via `PUT /api/v1/roles/{editor_id}/parent` with `{ parent_role_id: <viewer_id> }`
5. Create user U1, assign only "Editor" role (no direct "Viewer" assignment)
6. Call `POST /api/v1/check` with subject_id: U1, resource: "order", action: "read"
7. Assert `allowed === true` (inherited from parent "Viewer")
8. Call `POST /api/v1/check` with subject_id: U1, resource: "order", action: "create"
9. Assert `allowed === true` (direct from "Editor")

**Expected**: Child role inherits parent permissions; check evaluates both direct and inherited.

**Links to**: FR-007

---

## TC-005: Audit Log Created on Check

**Precondition**: Check Permission API is live; Audit Store is configured; audit writer is enabled.

**Steps**:
1. Perform a permission check (Allow or Deny) via `POST /api/v1/check`
2. Query `GET /api/v1/audit/events` with filters: date_from, date_to, event_type: "check", subject_id
3. Assert at least one event with event_type "check", matching subject_id, resource, action, decision

**Expected**: Every check (Allow or Deny) produces an audit log entry. Entry includes tenant_id, subject_id, resource, action, decision, timestamp.

**Links to**: FR-016

---

## TC-006: Multi-Tenancy Isolation

**Precondition**: Two tenants T1 and T2 exist. Each has users and roles. User U1 in T1 has `order:read`; T2 has no such assignment for U1 (or U1 does not exist in T2).

**Steps**:
1. Create tenant T1, role R1 with `order:read`, user U1, assign R1 to U1
2. Create tenant T2 (do not assign U1 any role in T2)
3. Call `POST /api/v1/check` with subject_id: U1, resource: "order", action: "read", `X-Tenant-ID: T2`
4. Assert `allowed === false` (U1 has no role in T2; tenant isolation)
5. Call `POST /api/v1/check` with subject_id: U1, resource: "order", action: "read", `X-Tenant-ID: T1`
6. Assert `allowed === true`
7. Call `GET /api/v1/roles` with `X-Tenant-ID: T2` — assert roles from T1 are not returned

**Expected**: No cross-tenant data. Check in T2 returns Deny when user has permissions only in T1. Admin list returns only tenant-scoped data.

**Links to**: FR-014

---

## TC-007: Check Permission — Cache Hit Performance

**Precondition**: PDP with Redis cache; same check repeated within TTL.

**Steps**:
1. Perform initial check (cache miss) — record latency L1
2. Perform same check again within 60s (cache hit) — record latency L2
3. Repeat step 2 for 100 iterations
4. Compute p99 of L2

**Expected**: p99 latency on cache hit < 50ms (NFR).

**Links to**: FR-011, FR-023, FR-013

---

## TC-008: Batch Check — Order Preserved

**Precondition**: User with mixed permissions (e.g. order:read yes, order:delete no).

**Steps**:
1. Create tenant, user, role with `order:read` only
2. Call `POST /api/v1/check/batch` with `{ checks: [{ subject_id, resource: "order", action: "read" }, { subject_id, resource: "order", action: "delete" }] }`
3. Assert response `results[0].allowed === true`, `results[1].allowed === false`
4. Assert results array length matches input checks length; order preserved

**Expected**: Batch results in same order as input; each check independent.

**Links to**: FR-011

---

## TC-009: Scope Filter on Check

**Precondition**: User has role with scope (e.g. Admin in org A only).

**Steps**:
1. Create tenant T1
2. Create role "Org Admin" with `user:read`, `user:update`
3. Create user U1, assign "Org Admin" with scope `{ scope_type: "org", scope_id: "org-A" }`
4. Call check with scope `{ scope_type: "org", scope_id: "org-A" }` for `user:read` — assert Allow
5. Call check with scope `{ scope_type: "org", scope_id: "org-B" }` for `user:read` — assert Deny

**Expected**: Scope filter applied; Allow only when scope matches assignment.

**Links to**: FR-015

---

## TC-010: Create Role — Unique Name per Tenant

**Precondition**: Tenant T1 exists.

**Steps**:
1. Create role "Order Manager" in tenant T1 — assert 201
2. Attempt to create role "Order Manager" again in tenant T1 — assert 400
3. Create tenant T2, create role "Order Manager" in T2 — assert 201 (same name allowed in different tenant)

**Expected**: Role name unique per tenant; duplicate returns 400; cross-tenant same name allowed.

**Links to**: FR-001

---

## TC-011: Update Role

**Precondition**: Role exists in tenant.

**Steps**:
1. Create role "Draft" with description "Initial"
2. Call `PUT /api/v1/roles/{id}` with `{ name: "Draft", description: "Updated", status: "active" }`
3. Assert 200; assert description is "Updated"
4. Call `GET /api/v1/roles/{id}` — assert updated fields

**Expected**: Role updated; audit event recorded.

**Links to**: FR-002

---

## TC-012: Delete Role

**Precondition**: Role exists; optional: role has permission assignments.

**Steps**:
1. Create role "ToDelete" with permissions
2. Call `DELETE /api/v1/roles/{id}`
3. Assert 204
4. Call `GET /api/v1/roles/{id}` — assert 404
5. Verify role_permissions cascade-deleted (or handled per policy)

**Expected**: Role deleted; related assignments cleared or handled; audit recorded.

**Links to**: FR-003

---

## TC-013: Assign Permissions to Role (PUT, PATCH, DELETE)

**Precondition**: Role exists.

**Steps**:
1. Create role "Test Role"
2. Call `PUT /api/v1/roles/{id}/permissions` with `{ permissions: ["order:read", "order:create"] }` — assert 200
3. Call `GET /api/v1/roles/{id}/permissions` — assert ["order:read", "order:create"]
4. Call `PATCH` with `{ permissions: ["order:update"] }` — assert 200
5. Call `GET` — assert all three permissions
6. Call `DELETE` with `{ permissions: ["order:create"] }` — assert 200
7. Call `GET` — assert ["order:read", "order:update"]

**Expected**: PUT replace, PATCH add, DELETE remove; wildcard supported per spec.

**Links to**: FR-004

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | QE | Initial test cases TC-001–TC-013 |
