package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grbac/grbac/internal/db"
	"github.com/grbac/grbac/internal/middleware"
	"github.com/grbac/grbac/internal/service"
)

// mockRoleService mocks role operations for handler tests.
type mockRoleService struct {
	createRole func(ctx context.Context, in service.CreateRoleInput) (db.Role, error)
	getRole    func(ctx context.Context, tenantID, roleID string) (db.Role, error)
	listRoles  func(ctx context.Context, in service.ListRolesInput) (service.ListRolesResult, error)
	updateRole func(ctx context.Context, in service.UpdateRoleInput) (db.Role, error)
	deleteRole func(ctx context.Context, tenantID, roleID string, actorID string) error
}

func (m *mockRoleService) CreateRole(ctx context.Context, in service.CreateRoleInput) (db.Role, error) {
	if m.createRole != nil {
		return m.createRole(ctx, in)
	}
	return db.Role{}, nil
}

func (m *mockRoleService) GetRole(ctx context.Context, tenantID, roleID string) (db.Role, error) {
	if m.getRole != nil {
		return m.getRole(ctx, tenantID, roleID)
	}
	return db.Role{}, nil
}

func (m *mockRoleService) ListRoles(ctx context.Context, in service.ListRolesInput) (service.ListRolesResult, error) {
	if m.listRoles != nil {
		return m.listRoles(ctx, in)
	}
	return service.ListRolesResult{}, nil
}

func (m *mockRoleService) UpdateRole(ctx context.Context, in service.UpdateRoleInput) (db.Role, error) {
	if m.updateRole != nil {
		return m.updateRole(ctx, in)
	}
	return db.Role{}, nil
}

func (m *mockRoleService) DeleteRole(ctx context.Context, tenantID, roleID string, actorID string) error {
	if m.deleteRole != nil {
		return m.deleteRole(ctx, tenantID, roleID, actorID)
	}
	return nil
}

func roleWithIDs(tenantID, roleID uuid.UUID, name string) db.Role {
	var tID, rID pgtype.UUID
	tID.Bytes = tenantID
	tID.Valid = true
	rID.Bytes = roleID
	rID.Valid = true
	return db.Role{
		ID:        rID,
		TenantID:  tID,
		Name:      name,
		Status:    "active",
		IsBuiltin: false,
		CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}
}

func TestCreateRole_Success(t *testing.T) {
	tid := uuid.New()
	rid := uuid.New()

	h := NewRolesHandler(&mockRoleService{
		createRole: func(ctx context.Context, in service.CreateRoleInput) (db.Role, error) {
			assert.Equal(t, tid.String(), in.TenantID)
			assert.Equal(t, "Order Manager", in.Name)
			return roleWithIDs(tid, rid, in.Name), nil
		},
	})

	body := `{"name":"Order Manager","description":"Manages orders"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true
	req = req.WithContext(middleware.WithTenantID(req.Context(), pgTid))

	w := httptest.NewRecorder()
	h.CreateRole(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp RoleResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, rid.String(), resp.ID)
	assert.Equal(t, tid.String(), resp.TenantID)
	assert.Equal(t, "Order Manager", resp.Name)
	assert.Equal(t, "active", resp.Status)
}

func TestCreateRole_Conflict(t *testing.T) {
	tid := uuid.New()
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	h := NewRolesHandler(&mockRoleService{
		createRole: func(ctx context.Context, in service.CreateRoleInput) (db.Role, error) {
			return db.Role{}, service.ErrRoleNameExists
		},
	})

	body := `{"name":"Order Manager"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithTenantID(req.Context(), pgTid))

	w := httptest.NewRecorder()
	h.CreateRole(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
	var errBody ErrorBody
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errBody))
	assert.Equal(t, "CONFLICT", errBody.Error.Code)
}

func TestCreateRole_Unauthorized_NoTenant(t *testing.T) {
	h := NewRolesHandler(&mockRoleService{})

	body := `{"name":"Order Manager"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateRole(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetRole_NotFound(t *testing.T) {
	tid := uuid.New()
	rid := uuid.New()
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	h := NewRolesHandler(&mockRoleService{
		getRole: func(ctx context.Context, tenantID, roleID string) (db.Role, error) {
			return db.Role{}, service.ErrRoleNotFound
		},
	})
	r := chi.NewRouter()
	r.Get("/roles/{id}", h.GetRole)

	req := httptest.NewRequest(http.MethodGet, "/roles/"+rid.String(), nil)
	req = req.WithContext(middleware.WithTenantID(req.Context(), pgTid))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	var errBody ErrorBody
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errBody))
	assert.Equal(t, "ROLE_NOT_FOUND", errBody.Error.Code)
}

func TestDeleteRole_NoContent(t *testing.T) {
	tid := uuid.New()
	rid := uuid.New()
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	h := NewRolesHandler(&mockRoleService{
		deleteRole: func(ctx context.Context, tenantID, roleID string, actorID string) error {
			assert.Equal(t, tid.String(), tenantID)
			assert.Equal(t, rid.String(), roleID)
			return nil
		},
	})
	r := chi.NewRouter()
	r.Delete("/roles/{id}", h.DeleteRole)

	req := httptest.NewRequest(http.MethodDelete, "/roles/"+rid.String(), nil)
	req = req.WithContext(middleware.WithTenantID(req.Context(), pgTid))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}
