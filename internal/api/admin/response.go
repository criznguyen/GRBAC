// Package admin provides Admin API handlers for tenants, roles, permissions.
package admin

import (
	"encoding/json"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/grbac/grbac/internal/middleware"
)

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	//nolint:errcheck // response committed; encode failure has no recovery
	_ = json.NewEncoder(w).Encode(v)
}

// ErrorBody matches api-spec-admin.md §4 error format.
type ErrorBody struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
	} `json:"error"`
}

// respondError writes an error response in api-spec format.
func respondError(w http.ResponseWriter, r *http.Request, code, message string, status int) {
	reqID := chimiddleware.GetReqID(r.Context())
	if reqID == "" {
		reqID = "unknown"
	}
	respondJSON(w, status, ErrorBody{
		Error: struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"request_id"`
		}{
			Code:      code,
			Message:   message,
			RequestID: reqID,
		},
	})
}

// TenantIDFromContext returns tenant ID string from context (set by Tenant middleware).
func TenantIDFromContext(r *http.Request) (string, bool) {
	tid, ok := middleware.TenantID(r.Context())
	if !ok || !tid.Valid {
		return "", false
	}
	return uuid.UUID(tid.Bytes).String(), true
}

// ActorIDFromContext returns the subject ID from context (set by Auth middleware). Used as actor_id for audit.
func ActorIDFromContext(r *http.Request) string {
	return middleware.SubjectID(r.Context())
}
