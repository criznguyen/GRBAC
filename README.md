# GRBAC — RBAC for Microservices

Hệ thống **Role-Based Access Control** tập trung cho kiến trúc microservices: một service authorization, nhiều tenant, đủ tính năng để dùng ngay cho production (enterprise & SME).

## Tính năng (Slice 0/1)

- **Tenants** — Tạo tenant, lấy thông tin tenant
- **Roles** — CRUD role (tên unique theo tenant), list có phân trang
- **Permissions** — Gán permission cho role (định dạng `resource:action`, hỗ trợ wildcard)
- **Auth** — JWT (HS256/RS256), header `X-Tenant-ID`
- **Audit** — Ghi mọi thay đổi role/permission vào `audit_admin`

## Yêu cầu

- Go 1.22+
- PostgreSQL 15+

## Chạy nhanh

```bash
# Clone & vào thư mục
cd GRBAC

# Cài dependency
go mod tidy

# Migrate DB (cần DATABASE_URL)
export DATABASE_URL="postgres://user:pass@localhost:5432/grbac?sslmode=disable"
make migrate-up

# Build & chạy API
export JWT_SECRET="your-secret"
make build && make run
```

API lắng nghe mặc định tại `http://localhost:8080`.

## API (base `/api/v1`)

| Method | Path | Mô tả |
|--------|------|--------|
| GET | /health | Health check |
| POST | /tenants | Tạo tenant |
| GET | /tenants/:id | Lấy tenant |
| POST | /roles | Tạo role |
| GET | /roles | List roles (paginated) |
| GET | /roles/:id | Lấy role |
| PUT | /roles/:id | Cập nhật role |
| DELETE | /roles/:id | Xóa role |
| PUT | /roles/:id/permissions | Thay toàn bộ permissions |
| PATCH | /roles/:id/permissions | Thêm permissions |
| DELETE | /roles/:id/permissions | Xóa permissions |
| GET | /roles/:id/permissions | List permissions của role |

**Headers bắt buộc:** `Authorization: Bearer <JWT>`, `X-Tenant-ID: <tenant-uuid>` (trừ POST /tenants).

## Test

```bash
make test              # Unit tests (không cần Docker)
make test-integration  # Integration tests (cần Docker, testcontainers)
make test-all          # Unit + integration
make test-e2e          # E2E: Postgres → migrate → API → smoke (cần migrate, jq)
```

## Cấu trúc thư mục

```
cmd/api/           # Entry point API
cmd/gen-jwt/       # Công cụ tạo JWT cho smoke/E2E
internal/
  api/admin/       # Handlers: tenants, roles, permissions
  audit/           # Ghi audit_admin
  config/          # Load config từ env
  db/              # sqlc, migrations, pool
  middleware/      # Auth, Tenant
  service/         # Business logic
tests/integration/ # Integration tests (testcontainers)
scripts/           # smoke-test.sh, run-e2e.sh
docs/sdlc/         # PRD, FRS, kiến trúc, QE, dev handoff
```

## Tài liệu

- **Product & SDLC:** [docs/sdlc/](docs/sdlc/) — PRD, FRS, kiến trúc, API spec, test plan
- **Dev:** [docs/sdlc/DEV-COMPLETED-SLICE-01.md](docs/sdlc/DEV-COMPLETED-SLICE-01.md) — Tổng kết Slice 0/1, cách chạy
- **QE:** [docs/sdlc/qe/QE-AUTOMATION.md](docs/sdlc/qe/QE-AUTOMATION.md) — Test pyramid, CI

## Smoke test (server đang chạy)

```bash
export JWT_SECRET=your-secret
export SMOKE_TEST_TOKEN=$(go run ./cmd/gen-jwt)
./scripts/smoke-test.sh
```

## License

MIT (hoặc ghi rõ license bạn dùng).
