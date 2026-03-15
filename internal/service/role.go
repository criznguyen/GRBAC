// Package service contains business logic for tenants, roles, permissions.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/grbac/grbac/internal/audit"
	"github.com/grbac/grbac/internal/db"
)

// ErrRoleNotFound is returned when a role does not exist.
var ErrRoleNotFound = errors.New("role not found")

// ErrRoleNameExists is returned when a role name already exists for the tenant.
var ErrRoleNameExists = errors.New("role name already exists for this tenant")

const (
	// DefaultListLimit is the default page size for ListRoles.
	DefaultListLimit = 20
	// MaxListLimit is the maximum page size.
	MaxListLimit = 100
)

// RoleService provides role operations.
type RoleService struct {
	q     *db.Queries
	audit *audit.Writer
}

// NewRoleService creates a RoleService. auditWriter may be nil to skip audit logging.
func NewRoleService(q *db.Queries, auditWriter *audit.Writer) *RoleService {
	return &RoleService{q: q, audit: auditWriter}
}

// CreateRoleInput holds input for creating a role.
type CreateRoleInput struct {
	TenantID    string
	Name        string
	Description string
	Status      string
	IsBuiltin   bool
	ActorID     string // Subject from auth; used for audit if audit writer is set
}

// CreateRole creates a new role. Returns ErrRoleNameExists if name is duplicate per tenant.
// role_permissions cascade delete is handled by DB ON DELETE CASCADE.
func (s *RoleService) CreateRole(ctx context.Context, in CreateRoleInput) (db.Role, error) {
	tid, err := parseUUID(in.TenantID)
	if err != nil {
		return db.Role{}, ErrRoleNotFound
	}
	_, err = s.q.GetRoleByNameAndTenant(ctx, db.GetRoleByNameAndTenantParams{
		TenantID: tid,
		Name:     in.Name,
	})
	if err == nil {
		return db.Role{}, ErrRoleNameExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.Role{}, err
	}
	desc := pgtype.Text{}
	if in.Description != "" {
		desc.String = in.Description
		desc.Valid = true
	}
	status := in.Status
	if status == "" {
		status = "active"
	}
	role, err := s.q.CreateRole(ctx, db.CreateRoleParams{
		TenantID:    tid,
		Name:        in.Name,
		Description: desc,
		Status:      status,
		IsBuiltin:   in.IsBuiltin,
	})
	if err != nil {
		return db.Role{}, err
	}
	if s.audit != nil && in.ActorID != "" {
		if err := s.audit.AppendAdminEvent(ctx, in.ActorID, "role.create", "role", uuid.UUID(role.ID.Bytes).String(),
			fmt.Sprintf("created role %q", in.Name), tid); err != nil {
			slog.Warn("audit write failed", "action", "role.create", "error", err)
		}
	}
	return role, nil
}

// GetRole returns a role by ID within tenant.
func (s *RoleService) GetRole(ctx context.Context, tenantID, roleID string) (db.Role, error) {
	tid, err := parseUUID(tenantID)
	if err != nil {
		return db.Role{}, ErrRoleNotFound
	}
	rid, err := parseUUID(roleID)
	if err != nil {
		return db.Role{}, ErrRoleNotFound
	}
	r, err := s.q.GetRole(ctx, db.GetRoleParams{ID: rid, TenantID: tid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Role{}, ErrRoleNotFound
		}
		return db.Role{}, err
	}
	return r, nil
}

// ListRolesInput holds pagination params.
type ListRolesInput struct {
	TenantID string
	Limit    int32
	Cursor   string
}

// ListRolesResult holds paginated roles.
type ListRolesResult struct {
	Items      []db.Role
	NextCursor string
	HasMore    bool
}

// ListRoles returns paginated roles for a tenant.
func (s *RoleService) ListRoles(ctx context.Context, in ListRolesInput) (ListRolesResult, error) {
	tid, err := parseUUID(in.TenantID)
	if err != nil {
		return ListRolesResult{}, ErrRoleNotFound
	}
	limit := in.Limit
	if limit <= 0 {
		limit = DefaultListLimit
	}
	if limit > MaxListLimit {
		limit = MaxListLimit
	}
	cursor := pgtype.UUID{Bytes: [16]byte{}, Valid: true} // zero UUID for first page
	if in.Cursor != "" {
		c, parseErr := parseUUID(in.Cursor)
		if parseErr != nil {
			return ListRolesResult{}, ErrRoleNotFound
		}
		cursor = c
	}
	// Fetch limit+1 to detect has_more
	roles, err := s.q.ListRolesByTenant(ctx, db.ListRolesByTenantParams{
		TenantID: tid,
		ID:       cursor,
		Limit:    limit + 1,
	})
	if err != nil {
		return ListRolesResult{}, err
	}
	hasMore := len(roles) > int(limit)
	if hasMore {
		roles = roles[:limit]
	}
	var nextCursor string
	if len(roles) > 0 && hasMore {
		nextCursor = uuid.UUID(roles[len(roles)-1].ID.Bytes).String()
	}
	return ListRolesResult{
		Items:      roles,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// UpdateRoleInput holds input for updating a role.
type UpdateRoleInput struct {
	TenantID    string
	RoleID      string
	Name        string
	Description string
	Status      string
	ActorID     string // Subject from auth; used for audit if audit writer is set
}

// UpdateRole updates a role. Returns ErrRoleNotFound if not found.
// Unique name per tenant: if name changes, check for duplicate.
func (s *RoleService) UpdateRole(ctx context.Context, in UpdateRoleInput) (db.Role, error) {
	tid, err := parseUUID(in.TenantID)
	if err != nil {
		return db.Role{}, ErrRoleNotFound
	}
	rid, err := parseUUID(in.RoleID)
	if err != nil {
		return db.Role{}, ErrRoleNotFound
	}
	_, err = s.q.GetRole(ctx, db.GetRoleParams{ID: rid, TenantID: tid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Role{}, ErrRoleNotFound
		}
		return db.Role{}, err
	}
	// Check unique name if name is being changed
	if in.Name != "" {
		existing, getErr := s.q.GetRoleByNameAndTenant(ctx, db.GetRoleByNameAndTenantParams{
			TenantID: tid,
			Name:     in.Name,
		})
		if getErr == nil && existing.ID.Bytes != rid.Bytes {
			return db.Role{}, ErrRoleNameExists
		}
	}
	desc := pgtype.Text{}
	if in.Description != "" {
		desc.String = in.Description
		desc.Valid = true
	}
	status := in.Status
	if status == "" {
		status = "active"
	}
	role, err := s.q.UpdateRole(ctx, db.UpdateRoleParams{
		ID:          rid,
		Name:        in.Name,
		Description: desc,
		Status:      status,
		TenantID:    tid,
	})
	if err != nil {
		return db.Role{}, err
	}
	if s.audit != nil && in.ActorID != "" {
		if err := s.audit.AppendAdminEvent(ctx, in.ActorID, "role.update", "role", in.RoleID,
			fmt.Sprintf("updated role %q", in.Name), tid); err != nil {
			slog.Warn("audit write failed", "action", "role.update", "error", err)
		}
	}
	return role, nil
}

// DeleteRole deletes a role. role_permissions cascade is handled by DB.
// actorID is the subject from auth; used for audit if audit writer is set.
func (s *RoleService) DeleteRole(ctx context.Context, tenantID, roleID string, actorID string) error {
	role, err := s.GetRole(ctx, tenantID, roleID)
	if err != nil {
		return err
	}
	tid, err := parseUUID(tenantID)
	if err != nil {
		return err
	}
	rid, err := parseUUID(roleID)
	if err != nil {
		return err
	}
	if err := s.q.DeleteRole(ctx, db.DeleteRoleParams{ID: rid, TenantID: tid}); err != nil {
		return err
	}
	if s.audit != nil && actorID != "" {
		if err := s.audit.AppendAdminEvent(ctx, actorID, "role.delete", "role", roleID,
			fmt.Sprintf("deleted role %q", role.Name), tid); err != nil {
			slog.Warn("audit write failed", "action", "role.delete", "error", err)
		}
	}
	return nil
}

func parseUUID(s string) (pgtype.UUID, error) {
	uid, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	var pgUUID pgtype.UUID
	pgUUID.Bytes = uid
	pgUUID.Valid = true
	return pgUUID, nil
}
