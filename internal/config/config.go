// Package config loads application configuration from environment variables.
// Tech Lead decision: envconfig (12-factor, Kelsey Hightower).
package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration.
type Config struct {
	// Server
	ServerAddr string `envconfig:"SERVER_ADDR" default:":8080"`

	// Policy Store (PostgreSQL)
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	// Audit Store (PostgreSQL, optional — can be same as DATABASE_URL for SME)
	AuditDatabaseURL string `envconfig:"AUDIT_DATABASE_URL"`

	// Logging
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// JWT (for auth middleware)
	JWTSecret       string `envconfig:"JWT_SECRET"`        // HS256 secret; or leave empty if using JWKS
	JWKSURL         string `envconfig:"JWKS_URL"`          // OIDC JWKS endpoint
	JWTClaimsIssuer string `envconfig:"JWT_CLAIMS_ISSUER"` // Expected iss claim

	// Redis (optional for MVP)
	RedisURL string `envconfig:"REDIS_URL"`
}

// Load reads configuration from environment.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("envconfig: %w", err)
	}

	// Audit DB defaults to Policy Store if not set (SME single-DB scenario)
	if cfg.AuditDatabaseURL == "" {
		cfg.AuditDatabaseURL = cfg.DatabaseURL
	}

	return &cfg, nil
}
