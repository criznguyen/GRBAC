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

// mockTenantService mocks tenant operations for handler tests.
type mockTenantService struct {
	createTenant func(ctx context.Context, in service.CreateTenantInput) (db.Tenant, error)
	getTenant    func(ctx context.Context, id string) (db.Tenant, error)
}

func (m *mockTenantService) CreateTenant(ctx context.Context, in service.CreateTenantInput) (db.Tenant, error) {
	if m.createTenant != nil {
		return m.createTenant(ctx, in)
	}
	return db.Tenant{}, nil
}

func (m *mockTenantService) GetTenant(ctx context.Context, id string) (db.Tenant, error) {
	if m.getTenant != nil {
		return m.getTenant(ctx, id)
	}
	return db.Tenant{}, nil
}

func TestCreateTenant_Success(t *testing.T) {
	tid := uuid.New()
	now := time.Now().UTC()
	var pgUUID pgtype.UUID
	pgUUID.Bytes = tid
	pgUUID.Valid = true
	h := NewTenantsHandler(&mockTenantService{
		createTenant: func(ctx context.Context, in service.CreateTenantInput) (db.Tenant, error) {
			assert.Equal(t, "Acme Corp", in.Name)
			return db.Tenant{ID: pgUUID, Name: in.Name, CreatedAt: pgtype.Timestamptz{Time: now, Valid: true}}, nil
		},
	})

	body := `{"name":"Acme Corp"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateTenant(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp TenantResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, tid.String(), resp.ID)
	assert.Equal(t, "Acme Corp", resp.Name)
	assert.NotEmpty(t, resp.CreatedAt)
}

func TestCreateTenant_InvalidJSON(t *testing.T) {
	h := NewTenantsHandler(&mockTenantService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewReader([]byte(`{invalid`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateTenant(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var errBody ErrorBody
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errBody))
	assert.Equal(t, "BAD_REQUEST", errBody.Error.Code)
}

func TestCreateTenant_ValidationError(t *testing.T) {
	h := NewTenantsHandler(&mockTenantService{})

	body := `{"name":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateTenant(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var errBody ErrorBody
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errBody))
	assert.Equal(t, "VALIDATION_ERROR", errBody.Error.Code)
}

func TestGetTenant_Forbidden_NoTenantContext(t *testing.T) {
	h := NewTenantsHandler(&mockTenantService{})
	r := chi.NewRouter()
	r.Get("/tenants/{id}", h.GetTenant)

	req := httptest.NewRequest(http.MethodGet, "/tenants/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetTenant_Forbidden_TenantMismatch(t *testing.T) {
	tid := uuid.New()
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	h := NewTenantsHandler(&mockTenantService{})
	r := chi.NewRouter()
	r.Get("/tenants/{id}", h.GetTenant)

	req := httptest.NewRequest(http.MethodGet, "/tenants/"+uuid.New().String(), nil)
	req = req.WithContext(middleware.WithTenantID(req.Context(), pgTid))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetTenant_Success(t *testing.T) {
	tid := uuid.New()
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true
	now := time.Now().UTC()

	h := NewTenantsHandler(&mockTenantService{
		getTenant: func(ctx context.Context, id string) (db.Tenant, error) {
			assert.Equal(t, tid.String(), id)
			return db.Tenant{
				ID:        pgTid,
				Name:      "Acme Corp",
				CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			}, nil
		},
	})
	r := chi.NewRouter()
	r.Get("/tenants/{id}", h.GetTenant)

	req := httptest.NewRequest(http.MethodGet, "/tenants/"+tid.String(), nil)
	req = req.WithContext(middleware.WithTenantID(req.Context(), pgTid))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp TenantResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, tid.String(), resp.ID)
	assert.Equal(t, "Acme Corp", resp.Name)
}

func TestGetTenant_NotFound(t *testing.T) {
	tid := uuid.New()
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	h := NewTenantsHandler(&mockTenantService{
		getTenant: func(ctx context.Context, id string) (db.Tenant, error) {
			return db.Tenant{}, service.ErrTenantNotFound
		},
	})
	r := chi.NewRouter()
	r.Get("/tenants/{id}", h.GetTenant)

	req := httptest.NewRequest(http.MethodGet, "/tenants/"+tid.String(), nil)
	req = req.WithContext(middleware.WithTenantID(req.Context(), pgTid))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	var errBody ErrorBody
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errBody))
	assert.Equal(t, "TENANT_NOT_FOUND", errBody.Error.Code)
}
