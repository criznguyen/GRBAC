# QE Automation — Test Pyramid, CI, Traceability

**Version:** 1.0  
**Date:** 2025-03-15  
**Scope:** Slice 0/1 (Policy Store + Admin API)  
**Status:** Active

---

## 1. Test Pyramid

| Layer        | Purpose                    | How to run              | Docker required |
|-------------|----------------------------|-------------------------|-----------------|
| **Unit**    | Fast; services, middleware | `make test`             | No              |
| **Integration** | API + DB (testcontainers) | `make test-integration` | Yes             |
| **E2E**     | Real server + smoke flow   | `make test-e2e`         | Yes             |

- **Unit**: `go test ./...` (excludes `//go:build integration`). No DB, no network.
- **Integration**: `go test -tags=integration ./tests/integration/...`. Uses testcontainers (PostgreSQL); in-process HTTP server.
- **E2E**: Script `scripts/run-e2e.sh` — start Postgres (Docker), migrate, start API process, run `scripts/smoke-test.sh` with generated JWT, then cleanup.

---

## 2. Running Locally

### Prerequisites

- Go 1.22+
- **Unit only:** none
- **Integration:** Docker (for testcontainers)
- **E2E:** Docker, `migrate` CLI, `jq`, `curl`

Install migrate (for E2E):

```bash
make install-tools
# or: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Commands

| Goal              | Command                |
|-------------------|------------------------|
| Unit tests        | `make test`            |
| Integration tests | `make test-integration`|
| All automation    | `make test-all`        |
| E2E (full flow)   | `make test-e2e`        |
| Smoke only        | Server + JWT required; see below |

### Smoke test (manual)

With server already running (`make run`) and same `JWT_SECRET` as server:

```bash
# Generate JWT (requires JWT_SECRET)
export JWT_SECRET=your-secret
export SMOKE_TEST_TOKEN=$(go run ./cmd/gen-jwt)
# Or build once: make build-gen-jwt && SMOKE_TEST_TOKEN=$(./bin/gen-jwt)

BASE_URL=http://localhost:8080 ./scripts/smoke-test.sh
```

Or use env: `SUB`, `TENANT_ID`, `EXP_HOURS` (see `cmd/gen-jwt`).

---

## 3. CI (GitHub Actions)

**Workflow:** `.github/workflows/test.yml`

**Triggers:** Push and pull requests to `main` / `master`.

| Job               | What it does                          |
|-------------------|----------------------------------------|
| **unit-tests**    | `make test`                            |
| **integration-tests** | `make test-integration` (Docker available) |
| **build**         | `make build`; uploads `bin/grbac-api` as artifact |
| **e2e**           | Install migrate, then `./scripts/run-e2e.sh` |

All jobs use `actions/setup-go` with Go 1.22. Integration and E2E run on `ubuntu-latest` with Docker.

---

## 4. Traceability — Slice 0/1 Acceptance Criteria

Reference: **HANDOFF-TO-QE-SLICE-01** (`docs/sdlc/ba/technical/HANDOFF-TO-QE-SLICE-01.md`).

| AC   | Description                          | Covered by                          |
|------|--------------------------------------|-------------------------------------|
| **T1–T6** | Tenants (create, get, 404, X-Tenant-ID) | Integration: `TestIntegration_Tenants_*`; E2E/smoke: tenant create + get |
| **R1–R8** | Roles (CRUD, duplicate name, tenant isolation) | Integration: `TestIntegration_Roles_*`; E2E/smoke: create role, get |
| **P1–P6** | Permissions (PUT/PATCH/DELETE/GET, wildcards) | Integration: `TestIntegration_Permissions_*`; E2E/smoke: assign + list permissions |
| **I1–I3** | Tenant isolation                     | Integration: tenant-scoped lists and 404 for wrong tenant |
| **A1–A5** | Admin audit events                    | Integration: `TestIntegration_Audit_*` (audit_admin events) |

**Test cases (docs):** `docs/sdlc/qe/test-cases-RBAC.md` (TC-006, TC-010, TC-011, TC-012, TC-013, TC-019).  
**Test plan:** `docs/sdlc/qe/test-plan-RBAC.md`.

---

## 5. Artifacts and Scripts

| Item | Purpose |
|------|--------|
| `cmd/gen-jwt/main.go` | HS256 JWT for scripts; env: `JWT_SECRET` (required), `SUB`, `TENANT_ID`, `EXP_HOURS` |
| `scripts/smoke-test.sh` | Six-step API flow: health → tenant → role → permissions → get role permissions → get tenant |
| `scripts/run-e2e.sh` | E2E orchestration: Postgres container → migrate → API → gen JWT → smoke-test → cleanup |

---

## 6. References

- **Handoff to QE (Slice 01):** `docs/sdlc/ba/technical/HANDOFF-TO-QE-SLICE-01.md`
- **Test plan:** `docs/sdlc/qe/test-plan-RBAC.md`
- **Test cases:** `docs/sdlc/qe/test-cases-RBAC.md`
- **Implementation plan:** `docs/sdlc/IMPLEMENTATION-PLAN.md`
