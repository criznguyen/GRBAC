//go:build integration

package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"

	"github.com/grbac/grbac/internal/audit"
	"github.com/grbac/grbac/internal/db"
	"github.com/grbac/grbac/internal/middleware"
	"github.com/grbac/grbac/internal/service"

	admin "github.com/grbac/grbac/internal/api/admin"
)

const testJWTSecret = "test-secret"

// StartTestServer creates a test server with the given dbURL, using test JWT secret.
// Returns the httptest.Server and a cleanup function.
// Config is set via env (DATABASE_URL, AUDIT_DATABASE_URL, JWT_SECRET) for the duration of the test.
func StartTestServer(t *testing.T, dbURL string) (*httptest.Server, func()) {
	t.Helper()

	// Set env so config.Load() works when we need it; middleware uses JWT_SECRET.
	// For DB we pass dbURL directly to avoid env races between tests.
	prevDB := os.Getenv("DATABASE_URL")
	prevAudit := os.Getenv("AUDIT_DATABASE_URL")
	prevJWT := os.Getenv("JWT_SECRET")
	os.Setenv("DATABASE_URL", dbURL)
	os.Setenv("AUDIT_DATABASE_URL", dbURL)
	os.Setenv("JWT_SECRET", testJWTSecret)

	cleanup := func() {
		os.Setenv("DATABASE_URL", prevDB)
		os.Setenv("AUDIT_DATABASE_URL", prevAudit)
		os.Setenv("JWT_SECRET", prevJWT)
	}

	ctx := context.Background()
	pool, err := db.OpenPool(ctx, dbURL)
	require.NoError(t, err, "failed to open DB pool")

	queries := db.New(pool)
	auditWriter := audit.NewWriter(queries)
	tenantSvc := service.NewTenantService(queries)
	roleSvc := service.NewRoleService(queries, auditWriter)
	permSvc := service.NewPermissionService(pool, queries, auditWriter)

	tenantsHandler := admin.NewTenantsHandler(tenantSvc)
	rolesHandler := admin.NewRolesHandler(roleSvc)
	permissionsHandler := admin.NewPermissionsHandler(permSvc)

	authCfg := middleware.AuthConfig{
		JWTSecret: testJWTSecret,
	}

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.Auth(authCfg))
		r.Post("/tenants", tenantsHandler.CreateTenant)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Tenant(queries))
			r.Get("/tenants/{id}", tenantsHandler.GetTenant)
			r.Post("/roles", rolesHandler.CreateRole)
			r.Get("/roles", rolesHandler.ListRoles)
			r.Get("/roles/{id}", rolesHandler.GetRole)
			r.Put("/roles/{id}", rolesHandler.UpdateRole)
			r.Delete("/roles/{id}", rolesHandler.DeleteRole)
			r.Get("/roles/{id}/permissions", permissionsHandler.ListRolePermissions)
			r.Put("/roles/{id}/permissions", permissionsHandler.ReplaceRolePermissions)
			r.Patch("/roles/{id}/permissions", permissionsHandler.AddRolePermissions)
			r.Delete("/roles/{id}/permissions", permissionsHandler.RemoveRolePermissions)
		})
	})

	srv := httptest.NewServer(r)
	cleanupWithPool := func() {
		srv.Close()
		pool.Close()
		cleanup()
	}
	return srv, cleanupWithPool
}
