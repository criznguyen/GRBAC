// Package service contains business logic for tenants, roles, permissions.
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/grbac/grbac/internal/db"
)

// ErrTenantNotFound is returned when a tenant does not exist.
var ErrTenantNotFound = errors.New("tenant not found")

// TenantService provides tenant operations.
type TenantService struct {
	q *db.Queries
}

// NewTenantService creates a TenantService.
func NewTenantService(q *db.Queries) *TenantService {
	return &TenantService{q: q}
}

// CreateTenantInput holds input for creating a tenant.
type CreateTenantInput struct {
	Name string `json:"name"`
}

// CreateTenant creates a new tenant.
func (s *TenantService) CreateTenant(ctx context.Context, in CreateTenantInput) (db.Tenant, error) {
	return s.q.CreateTenant(ctx, in.Name)
}

// GetTenant returns a tenant by ID.
func (s *TenantService) GetTenant(ctx context.Context, id string) (db.Tenant, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return db.Tenant{}, ErrTenantNotFound
	}
	var pgUUID pgtype.UUID
	if scanErr := pgUUID.Scan(uid); scanErr != nil {
		return db.Tenant{}, ErrTenantNotFound
	}
	t, err := s.q.GetTenant(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Tenant{}, ErrTenantNotFound
		}
		return db.Tenant{}, err
	}

	return t, nil
}
