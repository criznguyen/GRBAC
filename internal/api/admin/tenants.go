// Package admin provides Admin API handlers for tenants, roles, permissions.
package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/grbac/grbac/internal/db"
	"github.com/grbac/grbac/internal/service"
)

// TenantService defines tenant operations (implemented by service.TenantService).
type TenantService interface {
	CreateTenant(ctx context.Context, in service.CreateTenantInput) (db.Tenant, error)
	GetTenant(ctx context.Context, id string) (db.Tenant, error)
}

// TenantsHandler handles tenant API endpoints.
type TenantsHandler struct {
	tenant TenantService
}

// NewTenantsHandler creates a TenantsHandler.
func NewTenantsHandler(tenant TenantService) *TenantsHandler {
	return &TenantsHandler{tenant: tenant}
}

// CreateTenantRequest matches api-spec-admin.md §3.1.
type CreateTenantRequest struct {
	Name               string `json:"name" validate:"required,min=1,max=255"`
	DefaultPolicy      string `json:"default_policy,omitempty"`
	CreateBuiltinRoles *bool  `json:"create_builtin_roles,omitempty"`
}

// TenantResponse matches api-spec-admin.md §3.1 response.
type TenantResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func dbTenantToResponse(t db.Tenant) TenantResponse {
	return TenantResponse{
		ID:        uuid.UUID(t.ID.Bytes).String(),
		Name:      t.Name,
		CreatedAt: t.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// CreateTenant handles POST /tenants.
func (h *TenantsHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "BAD_REQUEST", "invalid JSON body", http.StatusBadRequest)
		return
	}
	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		respondError(w, r, "VALIDATION_ERROR", err.Error(), http.StatusBadRequest)
		return
	}

	t, err := h.tenant.CreateTenant(r.Context(), service.CreateTenantInput{Name: req.Name})
	if err != nil {
		respondError(w, r, "INTERNAL_ERROR", "failed to create tenant", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusCreated, dbTenantToResponse(t))
}

// GetTenant handles GET /tenants/{tenant_id}.
// Requires Tenant middleware; path :id must match X-Tenant-ID for tenant isolation.
func (h *TenantsHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, r, "BAD_REQUEST", "missing tenant id", http.StatusBadRequest)
		return
	}
	tenantID, ok := TenantIDFromContext(r)
	if !ok || tenantID != id {
		respondError(w, r, "FORBIDDEN", "tenant access denied", http.StatusForbidden)
		return
	}

	t, err := h.tenant.GetTenant(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTenantNotFound) {
			respondError(w, r, "TENANT_NOT_FOUND", "tenant not found", http.StatusNotFound)
			return
		}
		respondError(w, r, "INTERNAL_ERROR", "failed to get tenant", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, dbTenantToResponse(t))
}
