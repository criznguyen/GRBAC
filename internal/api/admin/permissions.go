// Package admin provides Admin API handlers for tenants, roles, permissions.
package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/grbac/grbac/internal/service"
)

// PermissionService defines permission operations (implemented by service.PermissionService).
type PermissionService interface {
	ReplaceRolePermissions(ctx context.Context, tenantID, roleID string, permissions []string, actorID string) ([]string, error)
	AddRolePermissions(ctx context.Context, tenantID, roleID string, permissions []string, actorID string) ([]string, error)
	RemoveRolePermissions(ctx context.Context, tenantID, roleID string, permissions []string, actorID string) ([]string, error)
	ListRolePermissions(ctx context.Context, tenantID, roleID string) ([]string, error)
}

// PermissionsHandler handles permission API endpoints for roles.
type PermissionsHandler struct {
	perm PermissionService
}

// NewPermissionsHandler creates a PermissionsHandler.
func NewPermissionsHandler(perm PermissionService) *PermissionsHandler {
	return &PermissionsHandler{perm: perm}
}

// PermissionsRequest matches api-spec-admin.md §3.3.
type PermissionsRequest struct {
	Permissions []string `json:"permissions" validate:"required,min=1,dive,required"`
}

// PermissionsResponse matches api-spec-admin.md §3.3.
type PermissionsResponse struct {
	RoleID      string   `json:"role_id"`
	Permissions []string `json:"permissions"`
}

// ReplaceRolePermissions handles PUT /roles/{id}/permissions.
func (h *PermissionsHandler) ReplaceRolePermissions(w http.ResponseWriter, r *http.Request) {
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
	actorID := ActorIDFromContext(r)

	var req PermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "BAD_REQUEST", "invalid JSON body", http.StatusBadRequest)
		return
	}
	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		respondError(w, r, "VALIDATION_ERROR", err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.perm.ReplaceRolePermissions(r.Context(), tenantID, roleID, req.Permissions, actorID)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			respondError(w, r, "ROLE_NOT_FOUND", "role not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrPermissionInvalidFormat) {
			respondError(w, r, "BAD_REQUEST", "permission must be in format resource:action", http.StatusBadRequest)
			return
		}
		respondError(w, r, "INTERNAL_ERROR", "failed to replace role permissions", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, PermissionsResponse{
		RoleID:      roleID,
		Permissions: result,
	})
}

// AddRolePermissions handles PATCH /roles/{id}/permissions.
func (h *PermissionsHandler) AddRolePermissions(w http.ResponseWriter, r *http.Request) {
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
	actorID := ActorIDFromContext(r)

	var req PermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "BAD_REQUEST", "invalid JSON body", http.StatusBadRequest)
		return
	}
	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		respondError(w, r, "VALIDATION_ERROR", err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.perm.AddRolePermissions(r.Context(), tenantID, roleID, req.Permissions, actorID)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			respondError(w, r, "ROLE_NOT_FOUND", "role not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrPermissionInvalidFormat) {
			respondError(w, r, "BAD_REQUEST", "permission must be in format resource:action", http.StatusBadRequest)
			return
		}
		respondError(w, r, "INTERNAL_ERROR", "failed to add role permissions", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, PermissionsResponse{
		RoleID:      roleID,
		Permissions: result,
	})
}

// RemoveRolePermissions handles DELETE /roles/{id}/permissions.
func (h *PermissionsHandler) RemoveRolePermissions(w http.ResponseWriter, r *http.Request) {
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
	actorID := ActorIDFromContext(r)

	var req PermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "BAD_REQUEST", "invalid JSON body", http.StatusBadRequest)
		return
	}
	if len(req.Permissions) == 0 {
		respondError(w, r, "BAD_REQUEST", "permissions array is required and cannot be empty", http.StatusBadRequest)
		return
	}
	validate := validator.New()
	for _, p := range req.Permissions {
		if err := validate.Var(p, "required"); err != nil {
			respondError(w, r, "VALIDATION_ERROR", "each permission must be non-empty", http.StatusBadRequest)
			return
		}
	}

	result, err := h.perm.RemoveRolePermissions(r.Context(), tenantID, roleID, req.Permissions, actorID)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			respondError(w, r, "ROLE_NOT_FOUND", "role not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrPermissionInvalidFormat) {
			respondError(w, r, "BAD_REQUEST", "permission must be in format resource:action", http.StatusBadRequest)
			return
		}
		respondError(w, r, "INTERNAL_ERROR", "failed to remove role permissions", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, PermissionsResponse{
		RoleID:      roleID,
		Permissions: result,
	})
}

// ListRolePermissions handles GET /roles/{id}/permissions.
func (h *PermissionsHandler) ListRolePermissions(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.perm.ListRolePermissions(r.Context(), tenantID, roleID)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			respondError(w, r, "ROLE_NOT_FOUND", "role not found", http.StatusNotFound)
			return
		}
		respondError(w, r, "INTERNAL_ERROR", "failed to list role permissions", http.StatusInternalServerError)
		return
	}
	if result == nil {
		result = []string{}
	}
	respondJSON(w, http.StatusOK, PermissionsResponse{
		RoleID:      roleID,
		Permissions: result,
	})
}
