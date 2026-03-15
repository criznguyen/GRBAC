package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grbac/grbac/internal/db"
)

type mockTenantLookup struct {
	err    error
	tenant db.Tenant
}

func (m *mockTenantLookup) GetTenant(ctx context.Context, id pgtype.UUID) (db.Tenant, error) {
	if m.err != nil {
		return db.Tenant{}, m.err
	}
	return m.tenant, nil
}

func TestTenant_MissingTenant(t *testing.T) {
	mock := &mockTenantLookup{}
	tenantMw := Tenant(mock)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	tenantMw(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing tenant")
}

func TestTenant_InvalidUUID(t *testing.T) {
	mock := &mockTenantLookup{}
	tenantMw := Tenant(mock)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(headerTenantID, "not-a-uuid")
	rec := httptest.NewRecorder()
	tenantMw(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid tenant")
}

func TestTenant_NotFound(t *testing.T) {
	mock := &mockTenantLookup{err: pgx.ErrNoRows}
	tenantMw := Tenant(mock)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called")
	})

	tid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(headerTenantID, tid.String())
	rec := httptest.NewRecorder()
	tenantMw(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "tenant not found")
}

func TestTenant_Found(t *testing.T) {
	tid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	mock := &mockTenantLookup{
		tenant: db.Tenant{ID: pgTid, Name: "Acme", CreatedAt: pgtype.Timestamptz{}},
	}
	tenantMw := Tenant(mock)

	var capturedTid pgtype.UUID
	var found bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTid, found = TenantID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(headerTenantID, tid.String())
	rec := httptest.NewRecorder()
	tenantMw(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.True(t, found)
	assert.Equal(t, tid[:], capturedTid.Bytes[:])
}

func TestTenant_FromAuthContext(t *testing.T) {
	tid := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	mock := &mockTenantLookup{
		tenant: db.Tenant{ID: pgTid, Name: "Acme", CreatedAt: pgtype.Timestamptz{}},
	}
	tenantMw := Tenant(mock)

	var capturedTid pgtype.UUID
	var found bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTid, found = TenantID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), contextKeyTenantClaim, tid.String()))
	rec := httptest.NewRecorder()
	tenantMw(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.True(t, found)
	assert.Equal(t, tid[:], capturedTid.Bytes[:])
}
