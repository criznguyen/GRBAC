# Technology Stack — RBAC for Microservices

**Version:** 1.0  
**Date:** 2025-03-15  
**Author:** Architect  
**Source:** ADRs 001–005, HANDOFF-TO-ARCHITECT.md

---

## 1. Overview

This document records the technology choices for the RBAC service. Choices align with NFRs (p99 < 50ms, 99.9% availability, audit compliance) and with ADRs (PDP topology, data store, cache, audit, multi-tenancy).

---

## 2. Summary Table

| Layer | Technology | Rationale |
|-------|------------|-----------|
| **Language / runtime** | **Go 1.22+** | Single binary, low memory, p99 < 50ms dễ đạt; scale-down (SME) và scale-up (Enterprise) tốt. See [tech-recommendation-enterprise-sme.md](./tech-recommendation-enterprise-sme.md). |
| **API / framework** | REST + optional gRPC; **Chi** or **Echo** | REST for Admin API and Check Permission; gRPC optional for low-latency check from SDK. Chi/Echo nhẹ, phù hợp cả SME và Enterprise. |
| **Policy store** | PostgreSQL | ADR-002: ACID, replication, partitioning; tenant_id on all tables. |
| **Audit store** | PostgreSQL (separate DB) | ADR-002, ADR-004: append-only, partitioning by time, retention job. |
| **Cache** | Redis (+ in-memory fallback) | ADR-003: Redis shared cache; TTL + invalidation; HA via cluster or Sentinel. SME single-node: in-memory fallback khi không có Redis. |
| **Message queue (optional)** | Optional (e.g. Redis Streams, Kafka) | For async audit write and/or cache invalidation events if not in-process. Technical BA to specify. |
| **IdP integration** | OAuth2 / OIDC (JWT or introspection) | Validate token; read subject (sub) and optional tenant claim. Library per language. |
| **SDKs** | Node, Go, Java | FR-012: thin clients calling Check Permission API; optional local in-memory cache. |

---

## 3. Component-Level Choices

### 3.1 Admin API

- **Protocol:** REST (JSON).
- **Auth:** Validate JWT or OAuth2 token from IdP; extract subject and tenant; authorize caller (RBAC or admin role) for admin operations.
- **Persistence:** PostgreSQL (Policy Store) for all CRUD; trigger cache invalidation (Redis) on policy/assignment changes.
- **Audit:** Write admin action events to Audit Store (sync or async per ADR-004).

### 3.2 PDP (Policy Decision Point)

- **Protocol:** REST and/or gRPC for Check Permission (and batch check).
- **Cache:** Redis for decision cache (key: tenant, subject, resource, action, scope); TTL default 60s; invalidation on change (ADR-003).
- **Policy read:** On cache miss, read from Policy Store (PostgreSQL); optionally use read replica.
- **Audit:** Async write of every check to Audit Store (ADR-004).

### 3.3 Policy Store

- **Database:** PostgreSQL.
- **Schema:** Tables for tenants, roles, permissions, role_permission, user_role, group_role, groups, group_membership, delegation, role_hierarchy, resource_registry; all keyed by tenant_id.
- **Replication:** Primary + replicas for HA and read scaling.

### 3.4 Audit Store

- **Database:** PostgreSQL (separate instance or DB).
- **Schema:** Append-only tables; partitioning by time (e.g. month); indexes for query/export filters.
- **Retention:** Configurable (min 1 year); job to drop/archive old partitions.

### 3.5 Cache

- **Product:** Redis.
- **Usage:** Permission check result cache; optional “resolved permissions per subject” cache.
- **HA:** Redis Cluster or Sentinel as per ops standards.

### 3.6 Message queue (optional)

- **Use cases:** Async audit write from PDP; optional cache invalidation broadcast.
- **Options:** Redis Streams (simpler), or Kafka (if already in platform). Technical BA to decide and specify.

---

## 4. SDK Stack

| SDK | Language | Behavior |
|-----|----------|----------|
| Node | TypeScript/JavaScript | HTTP client to Check Permission API; optional in-memory cache with TTL; retry and timeout. |
| Go | Go | Same; suitable for server-side microservices. |
| Java | Java 11+ | Same; optional Spring-friendly integration. |

SDKs do not implement PDP logic; they call the central PDP API. Local cache is optional and best-effort (TTL-based).

---

## 5. Deployment and Operations

- **Containers:** Admin API and PDP deployable as containers (e.g. Docker); stateless, horizontally scaled.
- **Load balancer:** In front of Admin API and PDP for 99.9% availability.
- **Secrets:** Tenant context, DB credentials, Redis URL, IdP client config from secret store (e.g. env or vault).
- **Observability:** Logging, metrics (latency, error rate, cache hit rate), tracing (optional) for Check Permission path.

---

## 6. Out of Scope (Phase 1)

- **Authentication:** Handled by IdP; RBAC only consumes identity.
- **Admin UI:** Phase 2.
- **Full ABAC/ReBAC:** Future extension.

---

## 7. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | Architect | Initial tech stack from ADRs |
