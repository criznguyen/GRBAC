# Team Breakdown — RBAC for Microservices

**Version:** 1.0  
**Date:** 2025-03-15  
**Author:** Technical BA  
**Source:** HANDOFF-TO-TECHNICAL-BA.md, container-diagram.md  
**Status:** Ready for Dev

---

## 1. Overview

Work is split into implementable slices with clear dependencies. Teams can work in parallel where dependencies allow.

---

## 2. Team Structure

| Team | Scope | Dependencies |
|------|-------|--------------|
| **Backend** | Admin API, PDP, Audit writer, Policy Store, Redis integration | DB schema, API specs |
| **SDK** | Node, Go, Java client libraries | Check Permission API contract |
| **DevOps/Infra** | PostgreSQL, Redis, deployment, retention jobs | Schema, config |

---

## 3. Backend Team

### 3.1 Scope

- **Admin API:** All CRUD endpoints (roles, permissions, users, groups, delegation, resource registry, policy import)
- **PDP (Check Permission API):** Single check, batch check; REST + optional gRPC
- **Audit writer:** Async write of check events; sync/async for admin events
- **Policy Store:** All reads/writes; tenant validation
- **Cache:** Redis integration; invalidation on policy change

### 3.2 Slices (Implementation Order)

| Slice | Description | Deps |
|-------|-------------|------|
| 1 | DB schema migration (Policy Store + Audit Store) | — |
| 2 | Admin API core: tenants, roles, permissions, role_permissions | Slice 1 |
| 3 | Admin API: users, user_roles, groups, group_members, group_roles | Slice 2 |
| 4 | Role hierarchy (role_hierarchy table, set parent API) | Slice 2 |
| 5 | PDP: Check Permission (single + batch), Policy Store read, deny-by-default | Slice 2 |
| 6 | Redis cache: check result cache, TTL, key shape | Slice 5 |
| 7 | Cache invalidation: Admin API triggers on all policy/assignment changes | Slice 3, 6 |
| 8 | Delegation: create rules, enforce in assign/revoke flow | Slice 3 |
| 9 | Resource registry CRUD | Slice 2 |
| 10 | Policy import (YAML/JSON) | Slice 2, 3 |
| 11 | Audit async writer: in-process queue or Redis Streams, Audit Store append | Slice 1 |
| 12 | Audit API: query, export (filters, pagination) | Slice 11 |
| 13 | Tenant bootstrap: create built-in roles, default policy | Slice 2 |

### 3.3 Acceptance Criteria (Backend)

- All Admin API endpoints return correct status codes; tenant isolation enforced
- Check Permission returns Allow/Deny per evaluation logic; p99 < 50ms on cache hit
- 100% of checks and admin actions written to audit
- Cache invalidated on every relevant policy/assignment change

---

## 4. SDK Team

### 4.1 Scope

- Node.js/TypeScript SDK
- Go SDK
- Java SDK

### 4.2 Slices

| Slice | Description | Deps |
|-------|-------------|------|
| 1 | Node SDK: `check(userId, resource, action, options)`; HTTP client; retry, timeout | Check Permission API live |
| 2 | Go SDK: same contract | Check Permission API live |
| 3 | Java SDK: same contract | Check Permission API live |
| 4 | Optional: in-memory local cache (TTL 10s) per SDK | Slice 1–3 |
| 5 | Optional: gRPC client if gRPC endpoint exists | Backend gRPC |

### 4.3 Acceptance Criteria (SDK)

- Each SDK integrates in < 1 day per service (per PRD metric)
- Method signature aligns with API spec; tenant and scope passed via options
- Timeout and retry documented; errors surfaced clearly

---

## 5. DevOps/Infra Team

### 5.1 Scope

- PostgreSQL: Policy Store + Audit Store provisioning; replication
- Redis: Provisioning; HA (cluster or Sentinel)
- Deployment: Admin API, PDP containers; load balancer
- Retention job: Drop/archive audit partitions older than retention period

### 5.2 Slices

| Slice | Description | Deps |
|-------|-------------|------|
| 1 | PostgreSQL: Policy Store + Audit Store DBs; schema apply | DB schema |
| 2 | Redis: instance/cluster; connection config | — |
| 3 | Container images: Admin API, PDP | Backend ready |
| 4 | Deployment: K8s/ECS; load balancer; secrets (DB, Redis, IdP) | Slice 1–3 |
| 5 | Audit partition automation: create monthly partitions | Schema |
| 6 | Retention job: drop partitions older than N months | Slice 5 |
| 7 | Observability: metrics (latency, cache hit rate, error rate), logging | Slice 4 |

### 5.3 Acceptance Criteria (DevOps)

- 99.9% availability target supported by HA config
- Audit retention configurable; retention job runs on schedule
- Metrics available for p99 latency and cache hit rate

---

## 6. Dependency Graph (Simplified)

```
DB Schema (Backend Slice 1)
    ├─> Admin API core (Slice 2, 3, 4)
    ├─> PDP (Slice 5)
    └─> Audit writer (Slice 11)

PDP (Slice 5) ──> Redis cache (Slice 6) ──> Invalidation (Slice 7)

Check Permission API ──> SDK Team (all slices)

PostgreSQL + Redis ──> DevOps deployment
```

---

## 7. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | Technical BA | Initial team breakdown |
