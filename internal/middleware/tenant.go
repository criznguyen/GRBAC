// Package middleware provides HTTP middleware: Auth (Bearer JWT), tenant validation, etc.
package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/grbac/grbac/internal/db"
)

const (
	headerTenantID                = "X-Tenant-ID"
	contextKeyTenantID contextKey = "tenant_id"
)

// TenantLookuper looks up a tenant by ID. Implemented by db.Queries.
type TenantLookuper interface {
	GetTenant(ctx context.Context, id pgtype.UUID) (db.Tenant, error)
}

// Tenant returns middleware that validates tenant (from X-Tenant-ID or JWT tenant claim) exists in DB,
// then injects tenant_id into context. Requires Auth middleware to run first for tenant-from-token.
// Returns 401 if tenant is missing, 403 if tenant not found.
func Tenant(lookup TenantLookuper) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantIDStr := r.Header.Get(headerTenantID)
			if tenantIDStr == "" {
				tenantIDStr = TenantClaim(r.Context())
			}
			if tenantIDStr == "" {
				respondTenantError(w, r, "UNAUTHORIZED", "missing tenant: set X-Tenant-ID header or tenant claim", http.StatusUnauthorized)
				return
			}

			parsed, err := uuid.Parse(tenantIDStr)
			if err != nil {
				respondTenantError(w, r, "FORBIDDEN", "invalid tenant id format", http.StatusForbidden)
				return
			}

			var tid pgtype.UUID
			tid.Bytes = parsed
			tid.Valid = true

			_, err = lookup.GetTenant(r.Context(), tid)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					respondTenantError(w, r, "FORBIDDEN", "tenant not found", http.StatusForbidden)
					return
				}
				slog.Error("tenant lookup failed", "error", err)
				respondTenantError(w, r, "INTERNAL_ERROR", "tenant lookup failed", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyTenantID, tid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TenantID returns the tenant ID from the request context (set by Tenant middleware).
func TenantID(ctx context.Context) (pgtype.UUID, bool) {
	v, ok := ctx.Value(contextKeyTenantID).(pgtype.UUID)
	return v, ok
}

// WithTenantID returns ctx with tenant ID set. Used for testing handlers that require tenant context.
func WithTenantID(ctx context.Context, tid pgtype.UUID) context.Context {
	return context.WithValue(ctx, contextKeyTenantID, tid)
}

func respondTenantError(w http.ResponseWriter, r *http.Request, code, message string, status int) {
	reqID := middleware.GetReqID(r.Context())
	if reqID == "" {
		reqID = "unknown"
	}
	body := map[string]any{
		"error": map[string]string{
			"code":       code,
			"message":    message,
			"request_id": reqID,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("failed to encode error response", "error", err)
	}
}
