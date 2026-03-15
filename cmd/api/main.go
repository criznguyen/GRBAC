// Package main — RBAC Admin API + PDP entry point.
// Tech Lead decisions: Chi, slog, envconfig.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/grbac/grbac/internal/api/admin"
	"github.com/grbac/grbac/internal/audit"
	"github.com/grbac/grbac/internal/config"
	"github.com/grbac/grbac/internal/db"
	"github.com/grbac/grbac/internal/middleware"
	"github.com/grbac/grbac/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	level := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))

	pool, err := db.OpenPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := db.New(pool)
	auditWriter := audit.NewWriter(queries)
	tenantSvc := service.NewTenantService(queries)
	roleSvc := service.NewRoleService(queries, auditWriter)
	permSvc := service.NewPermissionService(pool, queries, auditWriter)

	tenantsHandler := admin.NewTenantsHandler(tenantSvc)
	rolesHandler := admin.NewRolesHandler(roleSvc)
	permissionsHandler := admin.NewPermissionsHandler(permSvc)

	authCfg := middleware.AuthConfigFrom(cfg)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck // health response best-effort
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Admin API v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.Auth(authCfg))

		// Tenants: POST does not require tenant context
		r.Post("/tenants", tenantsHandler.CreateTenant)

		// Tenants: GET requires tenant validation
		r.Group(func(r chi.Router) {
			r.Use(middleware.Tenant(queries))
			r.Get("/tenants/{id}", tenantsHandler.GetTenant)
		})

		// Roles: all require tenant context
		r.Group(func(r chi.Router) {
			r.Use(middleware.Tenant(queries))
			r.Post("/roles", rolesHandler.CreateRole)
			r.Get("/roles", rolesHandler.ListRoles)
			r.Get("/roles/{id}", rolesHandler.GetRole)
			r.Put("/roles/{id}", rolesHandler.UpdateRole)
			r.Delete("/roles/{id}", rolesHandler.DeleteRole)
			// Permissions sub-resource
			r.Get("/roles/{id}/permissions", permissionsHandler.ListRolePermissions)
			r.Put("/roles/{id}/permissions", permissionsHandler.ReplaceRolePermissions)
			r.Patch("/roles/{id}/permissions", permissionsHandler.AddRolePermissions)
			r.Delete("/roles/{id}/permissions", permissionsHandler.RemoveRolePermissions)
		})
	})

	slog.Info("starting API server", "addr", cfg.ServerAddr)
	if err := http.ListenAndServe(cfg.ServerAddr, r); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
