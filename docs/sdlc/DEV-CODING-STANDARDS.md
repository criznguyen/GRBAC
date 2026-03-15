# Coding Standards — RBAC for Microservices

**Version:** 1.0  
**Date:** 2025-03-15  
**Author:** Tech Lead  
**Status:** Mandatory for all contributions

---

## 1. Overview

This document defines Go style, error handling, middleware patterns, and conventions. All code must pass `golangci-lint` before merge.

---

## 2. General Go Style

### 2.1 Formatting

- Use `gofmt` or `goimports` — enforced by golangci-lint.
- Max line length: 120 characters (soft).
- Prefer short variable names in small scopes; descriptive in larger scopes.

### 2.2 Naming

| Kind | Convention | Example |
|------|------------|---------|
| Package | Lowercase, single word | `config`, `middleware`, `service` |
| Interface | Single-method: verb; multi-method: -er suffix | `Reader`, `TenantValidator` |
| Errors | Prefix with package or domain | `ErrTenantNotFound`, `config.ErrInvalid` |
| Constants | PascalCase or SCREAMING for exported | `MaxPageSize`, `DefaultTTL` |

### 2.3 Comments

- All exported types and functions must have a doc comment.
- Start with the name of the symbol: `// Load reads config from env.`
- Use `//` for TODOs with ticket ref: `// TODO(GRBAC-123): add retry`.

---

## 3. Error Handling

### 3.1 Always Check Errors

```go
// BAD
db.Query(ctx, "SELECT ...")

// GOOD
rows, err := db.Query(ctx, "SELECT ...")
if err != nil {
    return fmt.Errorf("query roles: %w", err)
}
defer rows.Close()
```

### 3.2 Wrap with Context

Use `%w` for wrapping; preserve error chain for `errors.Is` / `errors.As`:

```go
if err != nil {
    return fmt.Errorf("get role %s: %w", id, err)
}
```

### 3.3 Sentinel Errors

Define package-level sentinels for expected conditions:

```go
var ErrTenantNotFound = errors.New("tenant not found")
var ErrRoleNotFound = errors.New("role not found")
```

---

## 4. Middleware Pattern

### 4.1 Signature

```go
func AuthMiddleware(validator JWTValidator) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token, err := extractBearer(r)
            if err != nil {
                respondError(w, http.StatusUnauthorized, "invalid token")
                return
            }
            claims, err := validator.Validate(r.Context(), token)
            if err != nil {
                respondError(w, http.StatusUnauthorized, "unauthorized")
                return
            }
            ctx := context.WithValue(r.Context(), contextKeyClaims, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 4.2 Context Keys

Use a private type for context keys to avoid collisions:

```go
type contextKey string
const contextKeyTenantID contextKey = "tenant_id"
```

---

## 5. API Handlers

### 5.1 Structure

- Handlers live in `internal/api/admin/` (or `internal/api/pdp/`).
- Inject dependencies via constructor or struct fields.
- Use `httptest` for unit tests.

### 5.2 Response Helpers

```go
func respondJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, message string) {
    respondJSON(w, status, map[string]string{"error": message})
}
```

### 5.3 Error Response Format

Match api-spec-admin.md:

```json
{
  "error": {
    "code": "ROLE_NOT_FOUND",
    "message": "Role with id xyz not found",
    "request_id": "req-uuid"
  }
}
```

---

## 6. Database (sqlc)

### 6.1 Query Location

- SQL in `internal/db/queries/*.sql`.
- Generate with `make sqlc-generate`; never edit generated code.

### 6.2 Transactions

Use `db.BeginTx` for multi-statement operations; always `defer tx.Rollback()` and call `tx.Commit()` on success.

---

## 7. Logging

- Use `slog` (stdlib).
- Log at appropriate level: `Debug` for verbose, `Info` for normal flow, `Warn` for recoverable, `Error` for failures.
- Include request_id in structured fields when available.
- Never log secrets (tokens, passwords).

---

## 8. Testing

### 8.1 Table-Driven Tests

```go
func TestParsePermission(t *testing.T) {
    tests := []struct {
        input    string
        wantRes  string
        wantAct  string
        wantErr  bool
    }{
        {"order:read", "order", "read", false},
        {"invalid", "", "", true},
    }
    for _, tt := range tests {
        res, act, err := parsePermission(tt.input)
        if (err != nil) != tt.wantErr {
            t.Errorf("parsePermission(%q) err=%v", tt.input, err)
        }
        // ...
    }
}
```

### 8.2 Handler Tests

Use `httptest.NewRequest` and `httptest.NewRecorder`; mock dependencies with interfaces.

---

## 9. Dependencies

- Prefer stdlib over third-party.
- Add dependencies only with Tech Lead approval.
- Run `go mod tidy` before commit.

---

## 10. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | Tech Lead | Initial standards |
