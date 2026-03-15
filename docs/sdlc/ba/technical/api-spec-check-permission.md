# Check Permission API Specification — RBAC for Microservices

**Version:** 1.0  
**Date:** 2025-03-15  
**Author:** Technical BA  
**Source:** HANDOFF-TO-TECHNICAL-BA.md, FRS-RBAC.md, ADR-001, ADR-003  
**Status:** Ready for Dev

---

## 1. Overview

The Check Permission API is the PDP (Policy Decision Point) entry point. Microservices (PEPs) call it to determine if a subject is allowed to perform an action on a resource. Results are cached in Redis; audit is written asynchronously.

**Base Path:** `/api/v1`  
**Traceability:** FR-011, FR-012, FR-013, FR-016, FR-023, FR-024

---

## 2. Authentication & Tenant Context

| Header | Required | Description |
|--------|----------|-------------|
| `Authorization` | Yes* | Bearer token. Subject (`sub`) and optionally tenant from token. |
| `X-Tenant-ID` | Yes* | Tenant identifier. Mandatory if not in token. |
| `X-Request-ID` | No | Optional request ID for tracing and audit correlation. |

---

## 3. Single Check

### POST /check

**Purpose:** Check if a subject is allowed to perform an action on a resource.

**Request:**
```json
{
  "subject_id": "user-uuid-or-sub-claim",
  "resource": "order",
  "action": "read",
  "scope": {
    "scope_type": "org",
    "scope_id": "org-123"
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| subject_id | string | Yes | User or service account ID (e.g. `sub` from token) |
| resource | string | Yes | Resource type (e.g. `order`, `user`) |
| action | string | Yes | Action (e.g. `read`, `create`, `delete`) |
| scope | object | No | Optional scope filter (org, project). Omit for global. |

**Response 200:**
```json
{
  "allowed": true
}
```

**Response 400:** Missing required fields, invalid format.

**Response 401:** Missing or invalid tenant/token.

---

## 4. Batch Check

### POST /check/batch

**Purpose:** Evaluate multiple permission checks in one request.

**Request:**
```json
{
  "checks": [
    {
      "subject_id": "user-1",
      "resource": "order",
      "action": "read"
    },
    {
      "subject_id": "user-1",
      "resource": "order",
      "action": "delete"
    }
  ]
}
```

**Response 200:**
```json
{
  "results": [
    { "allowed": true },
    { "allowed": false }
  ]
}
```

Order of results matches order of input checks. Each check is independent; one failure does not abort others.

---

## 5. Evaluation Logic

1. **Resolve subject:** Load user and groups for subject (tenant-scoped).
2. **Resolve roles:** Direct user roles + group roles + inherited (parent) roles.
3. **Resolve permissions:** Union of all permissions from those roles.
4. **Evaluate:** (resource, action) matches any permission (exact or wildcard).
5. **Scope:** If scope provided, filter assignments by scope; otherwise global.
6. **Deny by default:** If no matching permission, return `allowed: false`.

---

## 6. Caching (ADR-003)

**Key shape:** `rbac:check:{tenant_id}:{subject_id}:{resource}:{action}:{scope_hash}`

- Scope hash: MD5 or similar of scope JSON; empty string for global.
- **TTL:** Configurable; default 60 seconds.
- **Invalidation:** Admin API triggers on policy/assignment changes (see Admin API spec).

---

## 7. Audit

Every check (Allow or Deny) is logged to Audit Store asynchronously. Fields: `tenant_id`, `subject_id`, `resource`, `action`, `decision`, `timestamp`, `request_id`, `scope`. Write is off critical path to meet p99 < 50ms (ADR-004).

---

## 8. gRPC (Optional)

For low-latency SDK integration:

```protobuf
service CheckPermission {
  rpc Check(CheckRequest) returns (CheckResponse);
  rpc CheckBatch(BatchCheckRequest) returns (BatchCheckResponse);
}

message CheckRequest {
  string subject_id = 1;
  string resource = 2;
  string action = 3;
  Scope scope = 4;
}

message Scope {
  string scope_type = 1;
  string scope_id = 2;
}

message CheckResponse {
  bool allowed = 1;
}

message BatchCheckRequest {
  repeated CheckRequest checks = 1;
}

message BatchCheckResponse {
  repeated CheckResponse results = 1;
}
```

---

## 9. SDK Contract Alignment

SDKs (Node, Go, Java) expose:

```
check(userId, resource, action, options?)
  - options.tenant: override tenant
  - options.scope: { scope_type, scope_id }
  - options.timeout: request timeout (default 5s)
```

SDKs call POST /check (or gRPC). Optional local in-memory cache with shorter TTL (e.g. 10s) per ADR-001.

---

## 10. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | Technical BA | Initial Check Permission API spec |
