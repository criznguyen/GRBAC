# Audit API Specification — RBAC for Microservices

**Version:** 1.0  
**Date:** 2025-03-15  
**Author:** Technical BA  
**Source:** HANDOFF-TO-TECHNICAL-BA.md, FRS-RBAC.md, ADR-004  
**Status:** Ready for Dev

---

## 1. Overview

The Audit API provides query and export of audit logs. Two event types: (1) permission check events, (2) admin action events. All operations are tenant-scoped and require Auditor or Admin role.

**Base Path:** `/api/v1/audit`  
**Traceability:** FR-016, FR-017, FR-018

---

## 2. Authentication & Authorization

- **Required:** Bearer token, `X-Tenant-ID`.
- **Authorization:** Caller must have Auditor role or equivalent for query/export.
- **403 Forbidden:** If caller not authorized.

---

## 3. Query Audit

### GET /audit/events

**Purpose:** Query audit events with filters and pagination.

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| date_from | ISO 8601 | Yes | Start of time range |
| date_to | ISO 8601 | Yes | End of time range |
| event_type | string | No | `check` \| `admin` — filter by type |
| subject_id | string | No | Filter check events by subject |
| resource | string | No | Filter check events by resource |
| action | string | No | Filter check events by action |
| actor_id | string | No | Filter admin events by actor |
| action_type | string | No | Filter admin events (e.g. `create_role`, `assign_role_to_user`) |
| limit | int | No | Page size (default 50, max 200) |
| cursor | string | No | Pagination cursor |

**Response 200:**
```json
{
  "items": [
    {
      "event_type": "check",
      "id": "event-uuid",
      "tenant_id": "tenant-uuid",
      "subject_id": "user-uuid",
      "resource": "order",
      "action": "read",
      "decision": "Allow",
      "scope": { "scope_type": "org", "scope_id": "org-123" },
      "timestamp": "2025-03-15T10:00:00Z",
      "request_id": "req-uuid"
    },
    {
      "event_type": "admin",
      "id": "event-uuid",
      "tenant_id": "tenant-uuid",
      "actor_id": "admin-uuid",
      "action_type": "assign_role_to_user",
      "target_type": "user_role",
      "target_id": "assignment-uuid",
      "change_summary": "Assigned role R to user U",
      "timestamp": "2025-03-15T09:55:00Z"
    }
  ],
  "next_cursor": "cursor-abc",
  "has_more": true
}
```

**Validation:**
- `date_from` / `date_to`: Max range configurable (e.g. 90 days per request).
- Return 400 if invalid range or limits exceeded.

---

## 4. Export Audit

### POST /audit/export

**Purpose:** Export audit events to CSV or JSON.

**Request:**
```json
{
  "date_from": "2025-01-01T00:00:00Z",
  "date_to": "2025-03-15T23:59:59Z",
  "event_type": "check",
  "subject_id": "user-uuid",
  "resource": "order",
  "action": "read",
  "format": "csv",
  "limit": 10000
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| date_from | string | Yes | Start of range |
| date_to | string | Yes | End of range |
| event_type | string | No | `check` \| `admin` |
| subject_id | string | No | Filter |
| resource | string | No | Filter |
| action | string | No | Filter |
| actor_id | string | No | Filter (admin events) |
| format | string | Yes | `csv` \| `json` |
| limit | int | No | Max rows (default 10000, max per config) |

**Response 200 (sync, small exports):**

- `Content-Type: text/csv` or `application/json`
- `Content-Disposition: attachment; filename="audit-export-2025-03-15.csv"`
- Body: CSV or JSON array

**Response 202 (async, large exports):**
```json
{
  "job_id": "export-job-uuid",
  "status": "processing",
  "message": "Export started. Poll GET /audit/export/{job_id} for status."
}
```

**Polling:** `GET /audit/export/{job_id}` returns status and download URL when complete.

---

## 5. Event Schemas

### Check Event (event_type: check)

| Field | Type | Description |
|-------|------|-------------|
| id | string | Event UUID |
| event_type | string | `check` |
| tenant_id | string | Tenant |
| subject_id | string | User/service account |
| resource | string | Resource type |
| action | string | Action |
| decision | string | `Allow` \| `Deny` |
| scope | object | Optional scope |
| timestamp | string | ISO 8601 |
| request_id | string | Optional request correlation |

### Admin Event (event_type: admin)

| Field | Type | Description |
|-------|------|-------------|
| id | string | Event UUID |
| event_type | string | `admin` |
| tenant_id | string | Tenant |
| actor_id | string | Who performed the action |
| action_type | string | e.g. `create_role`, `assign_role_to_user` |
| target_type | string | e.g. `role`, `user_role` |
| target_id | string | Affected entity ID |
| change_summary | string | Human-readable summary |
| timestamp | string | ISO 8601 |

---

## 6. Retention & Limits

- **Retention:** Configurable; minimum 1 year (ADR-004).
- **Date range:** Export/query only within retention window.
- **Max export size:** Configurable (e.g. 100k rows); larger exports use async job.

---

## 7. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | Technical BA | Initial Audit API spec |
