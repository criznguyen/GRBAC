# GRBAC — RBAC for Microservices

A central **Role-Based Access Control** service for microservices: one authorization service, multi-tenant, production-ready for both enterprise and SME.

## Features (Slice 0/1)

- **Tenants** — Create tenant, get tenant by ID
- **Roles** — CRUD roles (unique name per tenant), paginated list
- **Permissions** — Assign permissions to roles (`resource:action` format, wildcard supported)
- **Auth** — JWT (HS256/RS256), `X-Tenant-ID` header
- **Audit** — All role/permission changes written to `audit_admin`

## Requirements

- Go 1.22+
- PostgreSQL 15+

## Quick start

```bash
# Clone and enter the repo
cd GRBAC

# Install dependencies
go mod tidy

# Run migrations (set DATABASE_URL)
export DATABASE_URL="postgres://user:pass@localhost:5432/grbac?sslmode=disable"
make migrate-up

# Build and run the API
export JWT_SECRET="your-secret"
make build && make run
```

The API listens on `http://localhost:8080` by default.

## API (base `/api/v1`)

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| POST | /tenants | Create tenant |
| GET | /tenants/:id | Get tenant |
| POST | /roles | Create role |
| GET | /roles | List roles (paginated) |
| GET | /roles/:id | Get role |
| PUT | /roles/:id | Update role |
| DELETE | /roles/:id | Delete role |
| PUT | /roles/:id/permissions | Replace all permissions |
| PATCH | /roles/:id/permissions | Add permissions |
| DELETE | /roles/:id/permissions | Remove permissions |
| GET | /roles/:id/permissions | List role permissions |

**Required headers:** `Authorization: Bearer <JWT>`, `X-Tenant-ID: <tenant-uuid>` (except for POST /tenants).

## Testing

```bash
make test              # Unit tests (no Docker)
make test-integration  # Integration tests (Docker + testcontainers)
make test-all          # Unit + integration
make test-e2e          # E2E: Postgres → migrate → API → smoke (requires migrate, jq)
```

## Project structure

```
cmd/api/           # API entry point
cmd/gen-jwt/       # JWT generator for smoke/E2E
internal/
  api/admin/       # Handlers: tenants, roles, permissions
  audit/           # audit_admin writer
  config/          # Config from env
  db/              # sqlc, migrations, pool
  middleware/      # Auth, Tenant
  service/         # Business logic
tests/integration/ # Integration tests (testcontainers)
scripts/           # smoke-test.sh, run-e2e.sh
docs/sdlc/         # PRD, FRS, architecture, QE, dev handoff
```

## Documentation

- **Product & SDLC:** [docs/sdlc/](docs/sdlc/) — PRD, FRS, architecture, API spec, test plan
- **Dev:** [docs/sdlc/DEV-COMPLETED-SLICE-01.md](docs/sdlc/DEV-COMPLETED-SLICE-01.md) — Slice 0/1 summary and run guide
- **QE:** [docs/sdlc/qe/QE-AUTOMATION.md](docs/sdlc/qe/QE-AUTOMATION.md) — Test pyramid, CI

## Smoke test (with server running)

```bash
export JWT_SECRET=your-secret
export SMOKE_TEST_TOKEN=$(go run ./cmd/gen-jwt)
./scripts/smoke-test.sh
```

## License

MIT (or specify your license).
