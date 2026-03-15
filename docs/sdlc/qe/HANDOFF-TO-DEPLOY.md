# Handoff to Deploy — RBAC Production Readiness

**Version:** 1.0  
**Date:** 2025-03-15  
**From:** QE  
**To:** DevOps / Deploy Team  
**Input docs:** test-plan-RBAC.md, test-cases-RBAC.md, HANDOFF-TO-DEV.md  
**Status:** Pre-Production Checklist

---

## 1. Purpose

This document defines what must be verified before promoting RBAC to production. It includes smoke tests, performance baselines, and operational checks.

---

## 2. Pre-Production Verification Checklist

### 2.1 Smoke Tests (Must Pass)

| # | Smoke Test | Command / Action | Expected |
|---|------------|------------------|----------|
| S1 | Health check | `GET /health` or `/api/v1/health` | 200 OK |
| S2 | Policy Store connectivity | Service starts; no DB connection errors | Startup success |
| S3 | Audit Store connectivity | Audit writer can write | No write failures |
| S4 | Redis connectivity | PDP cache read/write | No cache errors |
| S5 | Create tenant | `POST /api/v1/tenants` with valid payload | 201, tenant id returned |
| S6 | Create role | `POST /api/v1/roles` with tenant context | 201, role id returned |
| S7 | Assign permissions | `PUT /api/v1/roles/{id}/permissions` | 200 |
| S8 | Check permission (Allow) | `POST /api/v1/check` with allowed subject/resource/action | 200, `allowed: true` |
| S9 | Check permission (Deny) | `POST /api/v1/check` with no matching permission | 200, `allowed: false` |
| S10 | Audit query | `GET /api/v1/audit/events` with date range | 200, items array (may be empty) |
| S11 | Auth required | Request without `Authorization` header | 401 |
| S12 | Tenant required | Request without `X-Tenant-ID` | 401 |

### 2.2 Critical Path Tests (TC-001–TC-006)

All of the following must pass in staging/pre-prod:

- TC-001: Check Permission — Allow
- TC-002: Check Permission — Deny
- TC-003: Assign Role to User
- TC-004: Role Inheritance
- TC-005: Audit Log Created on Check
- TC-006: Multi-Tenancy Isolation

---

## 3. Performance Baseline

| Metric | Target | How to Verify |
|--------|--------|---------------|
| Check Permission (cache hit) p99 | < 50ms | Load test with repeated same check; measure p99 |
| Check Permission (cache miss) p95 | < 200ms | Cold check; acceptable until cache warms |
| Admin API (create role) p95 | < 500ms | Typical CRUD latency |
| Audit query (50 results) p95 | < 1s | Filtered query with pagination |

**Recommended tools:** k6, Artillery, or wrk for load; Prometheus/Grafana for metrics.

---

## 4. Infrastructure Verification

| Component | Check | Notes |
|-----------|-------|-------|
| PostgreSQL (Policy Store) | HA (primary + replica); connection pooling | Per ADR-002 |
| PostgreSQL (Audit Store) | Separate DB; partitioning enabled | Monthly partitions |
| Redis | HA (cluster or Sentinel) | Per ADR-003 |
| Load balancer | In front of Admin API and PDP | Stateless; no sticky sessions |
| Secrets | DB credentials, Redis URL, IdP config | From vault/env; no hardcoding |
| Retention job | Audit partition drop/archive | Runs on schedule; min 1 year retention |

---

## 5. Configuration Checklist

| Config | Default | Description |
|--------|---------|-------------|
| Cache TTL | 60s | Permission check result cache |
| Audit retention | 12 months | Minimum per ADR-004 |
| IdP / JWT | — | Token validation endpoint; required claims |
| Tenant validation | — | Validate X-Tenant-ID exists |

---

## 6. Rollback Plan

1. **Database migrations:** Use reversible migrations where possible; document rollback SQL.
2. **API versioning:** `/api/v1` — maintain backward compatibility for v1.
3. **Feature flags:** Optional: gate new PDP/Admin features if phased rollout.
4. **Traffic:** Ensure load balancer can route back to previous deployment if needed.

---

## 7. Post-Deploy Verification

Within 24 hours of production deploy:

1. Run smoke tests S1–S12 against production (with prod tenant).
2. Confirm metrics: latency, error rate, cache hit rate.
3. Confirm audit events are being written (sample query).
4. No critical alerts; logs clean of DB/Redis connection errors.

---

## 8. References

- **Test Plan:** `docs/sdlc/qe/test-plan-RBAC.md`
- **Test Cases:** `docs/sdlc/qe/test-cases-RBAC.md`
- **Handoff to Dev:** `docs/sdlc/ba/technical/HANDOFF-TO-DEV.md`
- **Implementation Plan:** `docs/sdlc/IMPLEMENTATION-PLAN.md`

---

## 9. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | QE | Initial handoff to Deploy |
