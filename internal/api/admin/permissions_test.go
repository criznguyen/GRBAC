package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grbac/grbac/internal/middleware"
	"github.com/grbac/grbac/internal/service"
)

// mockPermissionService mocks permission operations for handler tests.
type mockPermissionService struct {
	replaceRolePermissions func(ctx context.Context, tenantID, roleID string, permissions []string, actorID string) ([]string, error)
	addRolePermissions     func(ctx context.Context, tenantID, roleID string, permissions []string, actorID string) ([]string, error)
	removeRolePermissions  func(ctx context.Context, tenantID, roleID string, permissions []string, actorID string) ([]string, error)
	listRolePermissions    func(ctx context.Context, tenantID, roleID string) ([]string, error)
}

func (m *mockPermissionService) ReplaceRolePermissions(
	ctx context.Context, tenantID, roleID string, permissions []string, actorID string,
) ([]string, error) {
	if m.replaceRolePermissions != nil {
		return m.replaceRolePermissions(ctx, tenantID, roleID, permissions, actorID)
	}
	return nil, nil
}

func (m *mockPermissionService) AddRolePermissions(
	ctx context.Context, tenantID, roleID string, permissions []string, actorID string,
) ([]string, error) {
	if m.addRolePermissions != nil {
		return m.addRolePermissions(ctx, tenantID, roleID, permissions, actorID)
	}
	return nil, nil
}

func (m *mockPermissionService) RemoveRolePermissions(
	ctx context.Context, tenantID, roleID string, permissions []string, actorID string,
) ([]string, error) {
	if m.removeRolePermissions != nil {
		return m.removeRolePermissions(ctx, tenantID, roleID, permissions, actorID)
	}
	return nil, nil
}

func (m *mockPermissionService) ListRolePermissions(ctx context.Context, tenantID, roleID string) ([]string, error) {
	if m.listRolePermissions != nil {
		return m.listRolePermissions(ctx, tenantID, roleID)
	}
	return nil, nil
}

func TestListRolePermissions_Success(t *testing.T) {
	tid := uuid.New()
	rid := uuid.New()
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	h := NewPermissionsHandler(&mockPermissionService{
		listRolePermissions: func(ctx context.Context, tenantID, roleID string) ([]string, error) {
			assert.Equal(t, tid.String(), tenantID)
			assert.Equal(t, rid.String(), roleID)
			return []string{"order:read", "order:create"}, nil
		},
	})
	r := chi.NewRouter()
	r.Get("/roles/{id}/permissions", h.ListRolePermissions)

	req := httptest.NewRequest(http.MethodGet, "/roles/"+rid.String()+"/permissions", nil)
	req = req.WithContext(middleware.WithTenantID(req.Context(), pgTid))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp PermissionsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, rid.String(), resp.RoleID)
	assert.Equal(t, []string{"order:read", "order:create"}, resp.Permissions)
}

func TestListRolePermissions_NotFound(t *testing.T) {
	tid := uuid.New()
	rid := uuid.New()
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	h := NewPermissionsHandler(&mockPermissionService{
		listRolePermissions: func(ctx context.Context, tenantID, roleID string) ([]string, error) {
			return nil, service.ErrRoleNotFound
		},
	})
	r := chi.NewRouter()
	r.Get("/roles/{id}/permissions", h.ListRolePermissions)

	req := httptest.NewRequest(http.MethodGet, "/roles/"+rid.String()+"/permissions", nil)
	req = req.WithContext(middleware.WithTenantID(req.Context(), pgTid))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	var errBody ErrorBody
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errBody))
	assert.Equal(t, "ROLE_NOT_FOUND", errBody.Error.Code)
}

func TestListRolePermissions_Unauthorized_NoTenant(t *testing.T) {
	h := NewPermissionsHandler(&mockPermissionService{})
	r := chi.NewRouter()
	r.Get("/roles/{id}/permissions", h.ListRolePermissions)

	req := httptest.NewRequest(http.MethodGet, "/roles/"+uuid.New().String()+"/permissions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestReplaceRolePermissions_Success(t *testing.T) {
	tid := uuid.New()
	rid := uuid.New()
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	h := NewPermissionsHandler(&mockPermissionService{
		replaceRolePermissions: func(ctx context.Context, tenantID, roleID string, permissions []string, actorID string) ([]string, error) {
			assert.Equal(t, tid.String(), tenantID)
			assert.Equal(t, rid.String(), roleID)
			assert.Equal(t, []string{"order:read", "order:create"}, permissions)
			return []string{"order:read", "order:create"}, nil
		},
	})
	r := chi.NewRouter()
	r.Put("/roles/{id}/permissions", h.ReplaceRolePermissions)

	body := `{"permissions":["order:read","order:create"]}`
	req := httptest.NewRequest(http.MethodPut, "/roles/"+rid.String()+"/permissions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithTenantID(req.Context(), pgTid))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp PermissionsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, rid.String(), resp.RoleID)
	assert.Equal(t, []string{"order:read", "order:create"}, resp.Permissions)
}

func TestReplaceRolePermissions_NotFound(t *testing.T) {
	tid := uuid.New()
	rid := uuid.New()
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	h := NewPermissionsHandler(&mockPermissionService{
		replaceRolePermissions: func(ctx context.Context, tenantID, roleID string, permissions []string, actorID string) ([]string, error) {
			return nil, service.ErrRoleNotFound
		},
	})
	r := chi.NewRouter()
	r.Put("/roles/{id}/permissions", h.ReplaceRolePermissions)

	body := `{"permissions":["order:read"]}`
	req := httptest.NewRequest(http.MethodPut, "/roles/"+rid.String()+"/permissions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithTenantID(req.Context(), pgTid))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestAddRolePermissions_Success(t *testing.T) {
	tid := uuid.New()
	rid := uuid.New()
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	h := NewPermissionsHandler(&mockPermissionService{
		addRolePermissions: func(ctx context.Context, tenantID, roleID string, permissions []string, actorID string) ([]string, error) {
			assert.Equal(t, tid.String(), tenantID)
			assert.Equal(t, rid.String(), roleID)
			return []string{"order:read"}, nil
		},
	})
	r := chi.NewRouter()
	r.Patch("/roles/{id}/permissions", h.AddRolePermissions)

	body := `{"permissions":["order:read"]}`
	req := httptest.NewRequest(http.MethodPatch, "/roles/"+rid.String()+"/permissions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithTenantID(req.Context(), pgTid))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp PermissionsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, rid.String(), resp.RoleID)
	assert.Equal(t, []string{"order:read"}, resp.Permissions)
}

func TestRemoveRolePermissions_Success(t *testing.T) {
	tid := uuid.New()
	rid := uuid.New()
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	h := NewPermissionsHandler(&mockPermissionService{
		removeRolePermissions: func(ctx context.Context, tenantID, roleID string, permissions []string, actorID string) ([]string, error) {
			assert.Equal(t, tid.String(), tenantID)
			assert.Equal(t, rid.String(), roleID)
			assert.Equal(t, []string{"order:read"}, permissions)
			return []string{"order:read"}, nil
		},
	})
	r := chi.NewRouter()
	r.Delete("/roles/{id}/permissions", h.RemoveRolePermissions)

	body := `{"permissions":["order:read"]}`
	req := httptest.NewRequest(http.MethodDelete, "/roles/"+rid.String()+"/permissions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithTenantID(req.Context(), pgTid))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp PermissionsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, rid.String(), resp.RoleID)
	assert.Equal(t, []string{"order:read"}, resp.Permissions)
}
