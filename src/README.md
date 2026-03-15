# RBAC Microservices — Source Code Structure

**Version:** 1.0  
**Date:** 2025-03-15  
**Implementation Plan:** `docs/sdlc/IMPLEMENTATION-PLAN.md`

---

## 1. Overview

This directory contains the RBAC service source code. The structure follows the Technical BA handoff and architecture (Admin API, PDP, Audit, Policy Store, Audit Store).

---

## 2. Directory Layout

| Directory | Purpose |
|-----------|---------|
| `api/` | REST API layer — Admin API routes (tenants, roles, permissions, users, groups, etc.) and PDP routes (/check) |
| `pdp/` | Policy Decision Point — evaluation logic, cache integration, Check Permission handler |
| `audit/` | Audit writer — async/sync write of check and admin events to Audit Store |
| `db/` | Database layer — Policy Store and Audit Store clients, migrations |
| `sdk/` | Client SDK packages — Node, Go, Java (thin clients calling Check Permission API) |
| `middleware/` | Express/Fastify middleware — auth (Bearer + JWT), tenant validation, error handling |
| `services/` | Business logic — tenant, role, permission, user, group, delegation services |
| `config/` | Configuration — env vars, DB URLs, Redis, IdP settings |

---

## 3. Tech Stack

- **Runtime:** Go 1.22+ (per IMPLEMENTATION-PLAN.md)
- **Framework:** Chi or Echo
- **Policy Store:** PostgreSQL
- **Audit Store:** PostgreSQL (separate DB)
- **Cache:** Redis (in-memory fallback for SME)
- **Auth:** JWT validation (Bearer token)

---

## 4. Slice 0/1 (MVP) Scope

Current implementation focus:

- `db/migrations/` — Policy Store + Audit Store schema
- `api/admin/` — Tenants, Roles, Permissions (role-permission assignment)
- `middleware/` — Auth, X-Tenant-ID
- `audit/` — Admin action logging

---

## 5. Base Paths

- **Admin API:** `/api/v1`
- **Check Permission:** `/api/v1/check`, `/api/v1/check/batch`
- **Audit:** `/api/v1/audit/events`, `/api/v1/audit/export`

---

## 6. References

- Handoff: `docs/sdlc/ba/technical/HANDOFF-TO-DEV.md`
- API specs: `docs/sdlc/ba/technical/api-spec-*.md`
- DB schema: `docs/sdlc/ba/technical/db-schema.md`
- Implementation plan: `docs/sdlc/IMPLEMENTATION-PLAN.md`
