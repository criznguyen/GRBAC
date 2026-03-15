# Admin API Specification — RBAC for Microservices

**Version:** 1.0  
**Date:** 2025-03-15  
**Author:** Technical BA  
**Source:** HANDOFF-TO-TECHNICAL-BA.md, FRS-RBAC.md, ADRs 001–005  
**Status:** Ready for Dev

---

## 1. Overview

The Admin API provides control-plane operations for RBAC: CRUD roles, permissions, users, groups; assign/revoke roles; delegation; resource registry; policy import. All endpoints are tenant-scoped and require authentication.

**Base Path:** `/api/v1`  
**Traceability:** FR-001–FR-010, FR-019–FR-022

---

## 2. Authentication & Tenant Context

| Header | Required | Description |
|--------|----------|-------------|
| `Authorization` | Yes | Bearer token (JWT or OAuth2). IdP validates; RBAC extracts `sub` (subject) and optional `tenant` claim. |
| `X-Tenant-ID` | Yes* | Tenant identifier. Mandatory if not in token. Reject with 401 if missing or invalid. |

**Error responses:**
- `401 Unauthorized`: Missing/invalid token or tenant
- `403 Forbidden`: Caller not authorized for this operation in this tenant

---

## 3. API Endpoints (OpenAPI-style)

### 3.1 Tenants

| Method | Path | Purpose | FR |
|--------|------|---------|-----|
| POST | `/tenants` | Create tenant (bootstrap with default policy) | FR-021, FR-022 |
| GET | `/tenants/{tenant_id}` | Get tenant | FR-014 |

#### POST /tenants

**Purpose:** Create a new tenant with default policy (deny-all or configurable template).

**Request:**
```json
{
  "name": "Acme Corp",
  "default_policy": "deny_all",
  "create_builtin_roles": true
}
```

**Response 201:**
```json
{
  "id": "tenant-uuid",
  "name": "Acme Corp",
  "created_at": "2025-03-15T10:00:00Z"
}
```

---

### 3.2 Roles

| Method | Path | Purpose | FR |
|--------|------|---------|-----|
| POST | `/roles` | Create role | FR-001 |
| GET | `/roles` | List roles (paginated) | FR-001 |
| GET | `/roles/{role_id}` | Get role | FR-001 |
| PUT | `/roles/{role_id}` | Update role | FR-002 |
| DELETE | `/roles/{role_id}` | Delete role | FR-003 |
| PUT | `/roles/{role_id}/parent` | Set parent role (hierarchy) | FR-007 |

#### POST /roles

**Purpose:** Create a new role with unique name per tenant.

**Request:**
```json
{
  "name": "Order Manager",
  "description": "Manages orders",
  "status": "active"
}
```

**Response 201:**
```json
{
  "id": "role-uuid",
  "tenant_id": "tenant-uuid",
  "name": "Order Manager",
  "description": "Manages orders",
  "status": "active",
  "is_builtin": false,
  "created_at": "2025-03-15T10:00:00Z"
}
```

#### PUT /roles/{role_id}/parent

**Purpose:** Set parent role for inheritance. No circular reference allowed.

**Request:**
```json
{
  "parent_role_id": "parent-role-uuid"
}
```

**Response 200:** Updated role with parent reference.

---

### 3.3 Permissions

| Method | Path | Purpose | FR |
|--------|------|---------|-----|
| PUT | `/roles/{role_id}/permissions` | Assign permissions to role (replace) | FR-004 |
| PATCH | `/roles/{role_id}/permissions` | Add permissions | FR-004 |
| DELETE | `/roles/{role_id}/permissions` | Remove permissions | FR-004 |
| GET | `/roles/{role_id}/permissions` | List role permissions | FR-004 |

#### PUT /roles/{role_id}/permissions

**Purpose:** Replace all permissions for a role. Format: `resource:action` or wildcard (`*`).

**Request:**
```json
{
  "permissions": [
    "order:read",
    "order:create",
    "order:*",
    "*:read"
  ]
}
```

**Response 200:**
```json
{
  "role_id": "role-uuid",
  "permissions": ["order:read", "order:create", "order:*"]
}
```

---

### 3.4 Users

| Method | Path | Purpose | FR |
|--------|------|---------|-----|
| POST | `/users` | Create/register user reference | FR-005 |
| GET | `/users` | List users (paginated) | FR-005 |
| GET | `/users/{user_id}` | Get user | FR-005 |
| POST | `/users/{user_id}/roles` | Assign roles to user | FR-005 |
| DELETE | `/users/{user_id}/roles` | Revoke roles from user | FR-005 |

#### POST /users/{user_id}/roles

**Purpose:** Assign one or more roles to a user. Optional scope.

**Request:**
```json
{
  "role_ids": ["role-uuid-1", "role-uuid-2"],
  "scope": {
    "scope_type": "org",
    "scope_id": "org-123"
  }
}
```

**Response 201/200:** Assignment confirmation.

---

### 3.5 Groups

| Method | Path | Purpose | FR |
|--------|------|---------|-----|
| POST | `/groups` | Create group | FR-008 |
| GET | `/groups` | List groups (paginated) | FR-008 |
| GET | `/groups/{group_id}` | Get group | FR-008 |
| PUT | `/groups/{group_id}` | Update group | FR-008 |
| DELETE | `/groups/{group_id}` | Delete group | FR-008 |
| POST | `/groups/{group_id}/members` | Add members to group | FR-009 |
| DELETE | `/groups/{group_id}/members` | Remove members from group | FR-009 |
| POST | `/groups/{group_id}/roles` | Assign roles to group | FR-006 |
| DELETE | `/groups/{group_id}/roles` | Revoke roles from group | FR-006 |

#### POST /groups/{group_id}/members

**Request:**
```json
{
  "user_ids": ["user-uuid-1", "user-uuid-2"]
}
```

---

### 3.6 Delegations

| Method | Path | Purpose | FR |
|--------|------|---------|-----|
| POST | `/delegations` | Create delegation rule | FR-010 |
| GET | `/delegations` | List delegation rules | FR-010 |
| PUT | `/delegations/{delegation_id}` | Update delegation | FR-010 |
| DELETE | `/delegations/{delegation_id}` | Delete delegation | FR-010 |

#### POST /delegations

**Purpose:** Authorize user A to assign/revoke role R for users in scope S.

**Request:**
```json
{
  "delegator_user_id": "user-a-uuid",
  "role_id": "role-uuid",
  "scope": {
    "scope_type": "group",
    "scope_id": "group-uuid"
  },
  "allow_assign": true,
  "allow_revoke": true
}
```

---

### 3.7 Resource Registry

| Method | Path | Purpose | FR |
|--------|------|---------|-----|
| POST | `/resources` | Register resource type and actions | FR-019 |
| GET | `/resources` | List registered resources | FR-019 |
| GET | `/resources/{resource_type}` | Get resource definition | FR-019 |
| PUT | `/resources/{resource_type}` | Update resource definition | FR-019 |

#### POST /resources

**Purpose:** Register resource type and its allowed actions.

**Request:**
```json
{
  "resource_type": "order",
  "actions": ["read", "create", "update", "delete", "approve"],
  "description": "Order management resource"
}
```

---

### 3.8 Policy Import

| Method | Path | Purpose | FR |
|--------|------|---------|-----|
| POST | `/policy/import` | Import policy from YAML/JSON | FR-020 |

#### POST /policy/import

**Purpose:** Import policy definitions (roles, permissions, assignments) from file.

**Request:** `Content-Type: application/json` or `application/yaml`

**Body (JSON example):**
```json
{
  "version": 1,
  "roles": [
    {
      "name": "Viewer",
      "description": "Read-only access",
      "permissions": ["order:read", "user:read"]
    }
  ],
  "assignments": [
    { "user_id": "user-1", "role": "Viewer" },
    { "group_id": "group-1", "role": "Editor" }
  ]
}
```

**Response 200:**
```json
{
  "import_id": "import-uuid",
  "roles_created": 2,
  "assignments_applied": 5,
  "dry_run": false
}
```

**Query params:** `?dry_run=true` for preview without applying.

---

## 4. Common Schemas

### Error Response (4xx/5xx)

```json
{
  "error": {
    "code": "ROLE_NOT_FOUND",
    "message": "Role with id xyz not found",
    "request_id": "req-uuid"
  }
}
```

### Pagination (Query params)

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| limit | int | 20 | Max items per page (max 100) |
| cursor | string | — | Pagination cursor for next page |

**List response:**
```json
{
  "items": [...],
  "next_cursor": "abc123",
  "has_more": true
}
```

---

## 5. Cache Invalidation (Internal)

Admin API triggers Redis invalidation on every policy/assignment change per ADR-003:

| Event | Invalidation |
|-------|--------------|
| Role update/delete | All subjects with that role |
| Assign/revoke user role | Subject user_id |
| Assign/revoke group role | All members of group |
| Group membership change | Affected user(s) |
| Delegation change | Delegator user |
| Policy import | Per affected subject/role |

---

## 6. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | Technical BA | Initial Admin API spec |
