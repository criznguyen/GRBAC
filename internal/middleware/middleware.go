// Package middleware provides HTTP middleware: Auth (Bearer JWT), tenant validation, etc.
// Work package (1): Auth middleware + tenant validation.
//
// Chain order: Auth → Tenant (tenant needs auth for token tenant claim).
package middleware

// Auth and Tenant are the primary middleware exports.
// Use: r.Use(middleware.Auth(cfg)); r.Use(middleware.Tenant(queries))
