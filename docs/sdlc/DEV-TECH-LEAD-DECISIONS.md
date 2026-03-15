# Tech Lead Decisions — RBAC for Microservices

**Version:** 1.0  
**Date:** 2025-03-15  
**Author:** Tech Lead  
**Input:** tech-stack.md, IMPLEMENTATION-PLAN.md, db-schema.md, api-spec-admin.md  
**Status:** Final

---

## 1. Overview

This document records the final technology decisions for the RBAC backend. All choices align with NFRs (p99 < 50ms, 99.9% availability, audit compliance) and are production-ready.

---

## 2. Final Decisions

### 2.1 Framework: **Chi**

| Option | Decision | Rationale |
|--------|----------|-----------|
| Chi vs Echo | **Chi** | Chi is lightweight, standard library–aligned (`net/http`), uses no reflection, has minimal magic. Middleware is explicit and composable. Echo is more feature-rich but adds abstraction layers we don't need. Chi fits the “simple, explicit, production” ethos. Both are performant; Chi has a smaller dependency footprint. |

**Dependency:** `github.com/go-chi/chi/v5`

---

### 2.2 DB Layer: **sqlc + golang-migrate**

| Option | Decision | Rationale |
|--------|----------|-----------|
| sqlc + golang-migrate vs Atlas | **sqlc + golang-migrate** | sqlc generates type-safe Go from SQL—no ORM magic, no N+1 risks. golang-migrate is battle-tested, uses plain SQL files, and supports up/down migrations. Atlas offers schema diffing and multi-DB support but adds complexity. For Policy Store + Audit Store (PostgreSQL only), golang-migrate is sufficient and simpler. |

**Dependencies:**
- `github.com/sqlc-dev/sqlc` (CLI, dev)
- `github.com/golang-migrate/migrate/v4`
- `github.com/jackc/pgx/v5` (driver)

---

### 2.3 JWT: **github.com/golang-jwt/jwt/v5**

| Option | Decision | Rationale |
|--------|----------|-----------|
| golang-jwt/jwt v5 vs go-auth0/jwt | **golang-jwt/jwt v5** | Community standard, widely adopted, no vendor lock-in. v5 uses `RegisteredClaims` and `MapClaims`; we extract `sub` and optional `tenant` claim. go-auth0/jwt is Auth0-specific; we may support multiple IdPs (Auth0, Keycloak, custom). golang-jwt is IdP-agnostic. |

**Dependency:** `github.com/golang-jwt/jwt/v5`

---

### 2.4 Validation: **go-playground/validator v10**

| Option | Decision | Rationale |
|--------|----------|-----------|
| go-playground/validator | **Confirmed** | Struct tags (`validate:"required,min=1,max=255"`), JSON binding, custom validators. Standard choice for request validation in Go APIs. |

**Dependency:** `github.com/go-playground/validator/v10`

---

### 2.5 Config: **envconfig**

| Option | Decision | Rationale |
|--------|----------|-----------|
| viper vs envconfig vs std lib | **envconfig** | 12-factor app: config from environment. envconfig (Kelsey Hightower) is minimal: struct tags map env vars to fields. No file parsing, no remote config. viper is powerful but overkill for env-only config. Std lib would require manual parsing. |

**Dependency:** `github.com/kelseyhightower/envconfig`

---

### 2.6 Logging: **slog (stdlib)**

| Option | Decision | Rationale |
|--------|----------|-----------|
| slog vs zerolog vs zap | **slog** | slog is in the standard library (Go 1.21+). Structured logging, levels, JSON handler. No external dependency. zerolog and zap are faster but add deps. For RBAC, log volume is moderate; slog is sufficient and keeps the dependency tree minimal. |

**Usage:** `log/slog` — no external package.

---

### 2.7 Testing: **testify + httptest**

| Option | Decision | Rationale |
|--------|----------|-----------|
| testify | **Confirmed** | `assert`, `require`, `suite` for readable tests. |
| httptest | **Confirmed** | `net/http/httptest` for handler testing. |

**Dependencies:** `github.com/stretchr/testify` (assert, require)

---

### 2.8 Linting: **golangci-lint**

| Option | Decision | Rationale |
|--------|----------|-----------|
| golangci-lint | **Confirmed** | Aggregates multiple linters (staticcheck, errcheck, govet, etc.). Config in `.golangci.yml`. |

---

## 3. Summary Table

| Layer | Choice | Package |
|-------|--------|---------|
| **Framework** | Chi | `github.com/go-chi/chi/v5` |
| **DB layer** | sqlc + golang-migrate | `github.com/sqlc-dev/sqlc`, `github.com/golang-migrate/migrate/v4` |
| **DB driver** | pgx v5 | `github.com/jackc/pgx/v5` |
| **JWT** | golang-jwt v5 | `github.com/golang-jwt/jwt/v5` |
| **Validation** | validator v10 | `github.com/go-playground/validator/v10` |
| **Config** | envconfig | `github.com/kelseyhightower/envconfig` |
| **Logging** | slog | `log/slog` (stdlib) |
| **Testing** | testify, httptest | `github.com/stretchr/testify`, `net/http/httptest` |
| **Linting** | golangci-lint | `.golangci.yml` |

---

## 4. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | Tech Lead | Initial decisions |
