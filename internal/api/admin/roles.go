// Package admin provides Admin API handlers for tenants, roles, permissions.
package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/grbac/grbac/internal/db"
	"github.com/grbac/grbac/internal/service"
)

// RoleService defines role operations (implemented by service.RoleService).
type RoleService interface {
	CreateRole(ctx context.Context, in service.CreateRoleInput) (db.Role, error)
	GetRole(ctx context.Context, tenantID, roleID string) (db.Role, error)
	ListRoles(ctx context.Context, in service.ListRolesInput) (service.ListRolesResult, error)
	UpdateRole(ctx context.Context, in service.UpdateRoleInput) (db.Role, error)
	DeleteRole(ctx context.Context, tenantID, roleID string, actorID string) error
}

// RolesHandler handles role API endpoints.
type RolesHandler struct {
	role RoleService
}

// NewRolesHandler creates a RolesHandler.
func NewRolesHandler(role RoleService) *RolesHandler {
	return &RolesHandler{role: role}
}

// CreateRoleRequest matches api-spec-admin.md §3.2.
type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

// RoleResponse matches api-spec-admin.md §3.2 response.
type RoleResponse struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	IsBuiltin   bool   `json:"is_builtin"`
	CreatedAt   string `json:"created_at"`
}

func dbRoleToResponse(r db.Role) RoleResponse {
	resp := RoleResponse{
		ID:        uuid.UUID(r.ID.Bytes).String(),
		TenantID:  uuid.UUID(r.TenantID.Bytes).String(),
		Name:      r.Name,
		Status:    r.Status,
		IsBuiltin: r.IsBuiltin,
		CreatedAt: r.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
	if r.Description.Valid {
		resp.Description = r.Description.String
	}
	return resp
}

// CreateRole handles POST /roles.
func (h *RolesHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := TenantIDFromContext(r)
	if !ok {
		respondError(w, r, "UNAUTHORIZED", "tenant context required", http.StatusUnauthorized)
		return
	}

	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "BAD_REQUEST", "invalid JSON body", http.StatusBadRequest)
		return
	}
	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		respondError(w, r, "VALIDATION_ERROR", err.Error(), http.StatusBadRequest)
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	role, err := h.role.CreateRole(r.Context(), service.CreateRoleInput{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Status:      status,
		IsBuiltin:   false,
		ActorID:     ActorIDFromContext(r),
	})
	if err != nil {
		if errors.Is(err, service.ErrRoleNameExists) {
			respondError(w, r, "CONFLICT", "role name already exists for this tenant", http.StatusConflict)
			return
		}
		respondError(w, r, "INTERNAL_ERROR", "failed to create role", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusCreated, dbRoleToResponse(role))
}

// GetRole handles GET /roles/{role_id}.
func (h *RolesHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := TenantIDFromContext(r)
	if !ok {
		respondError(w, r, "UNAUTHORIZED", "tenant context required", http.StatusUnauthorized)
		return
	}
	roleID := chi.URLParam(r, "id")
	if roleID == "" {
		respondError(w, r, "BAD_REQUEST", "missing role id", http.StatusBadRequest)
		return
	}

	role, err := h.role.GetRole(r.Context(), tenantID, roleID)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			respondError(w, r, "ROLE_NOT_FOUND", "role not found", http.StatusNotFound)
			return
		}
		respondError(w, r, "INTERNAL_ERROR", "failed to get role", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, dbRoleToResponse(role))
}

// ListRoles handles GET /roles (paginated).
func (h *RolesHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := TenantIDFromContext(r)
	if !ok {
		respondError(w, r, "UNAUTHORIZED", "tenant context required", http.StatusUnauthorized)
		return
	}

	limitVal, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	limit := limitVal
	if err != nil {
		limit = service.DefaultListLimit
	}
	if limit <= 0 {
		limit = service.DefaultListLimit
	}
	if limit > service.MaxListLimit {
		limit = service.MaxListLimit
	}
	cursor := r.URL.Query().Get("cursor")

	result, err := h.role.ListRoles(r.Context(), service.ListRolesInput{
		TenantID: tenantID,
		Limit:    int32(limit),
		Cursor:   cursor,
	})
	if err != nil {
		respondError(w, r, "INTERNAL_ERROR", "failed to list roles", http.StatusInternalServerError)
		return
	}

	items := make([]RoleResponse, len(result.Items))
	for i, ro := range result.Items {
		items[i] = dbRoleToResponse(ro)
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"items":       items,
		"next_cursor": result.NextCursor,
		"has_more":    result.HasMore,
	})
}

// UpdateRoleRequest for PUT /roles/{role_id}.
type UpdateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

// UpdateRole handles PUT /roles/{role_id}.
func (h *RolesHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := TenantIDFromContext(r)
	if !ok {
		respondError(w, r, "UNAUTHORIZED", "tenant context required", http.StatusUnauthorized)
		return
	}
	roleID := chi.URLParam(r, "id")
	if roleID == "" {
		respondError(w, r, "BAD_REQUEST", "missing role id", http.StatusBadRequest)
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "BAD_REQUEST", "invalid JSON body", http.StatusBadRequest)
		return
	}
	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		respondError(w, r, "VALIDATION_ERROR", err.Error(), http.StatusBadRequest)
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	role, err := h.role.UpdateRole(r.Context(), service.UpdateRoleInput{
		TenantID:    tenantID,
		RoleID:      roleID,
		Name:        req.Name,
		Description: req.Description,
		Status:      status,
		ActorID:     ActorIDFromContext(r),
	})
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			respondError(w, r, "ROLE_NOT_FOUND", "role not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrRoleNameExists) {
			respondError(w, r, "CONFLICT", "role name already exists for this tenant", http.StatusConflict)
			return
		}
		respondError(w, r, "INTERNAL_ERROR", "failed to update role", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, dbRoleToResponse(role))
}

// DeleteRole handles DELETE /roles/{role_id}.
func (h *RolesHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := TenantIDFromContext(r)
	if !ok {
		respondError(w, r, "UNAUTHORIZED", "tenant context required", http.StatusUnauthorized)
		return
	}
	roleID := chi.URLParam(r, "id")
	if roleID == "" {
		respondError(w, r, "BAD_REQUEST", "missing role id", http.StatusBadRequest)
		return
	}

	err := h.role.DeleteRole(r.Context(), tenantID, roleID, ActorIDFromContext(r))
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			respondError(w, r, "ROLE_NOT_FOUND", "role not found", http.StatusNotFound)
			return
		}
		respondError(w, r, "INTERNAL_ERROR", "failed to delete role", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
