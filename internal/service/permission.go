// Package service contains business logic for tenants, roles, permissions.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grbac/grbac/internal/audit"
	"github.com/grbac/grbac/internal/db"
)

// ErrPermissionInvalidFormat is returned when a permission string is not in resource:action format.
var ErrPermissionInvalidFormat = errors.New("permission must be in format resource:action")

// PermissionService provides permission operations for roles.
type PermissionService struct {
	q     *db.Queries
	pool  *pgxpool.Pool
	audit *audit.Writer
}

// NewPermissionService creates a PermissionService.
func NewPermissionService(pool *pgxpool.Pool, q *db.Queries, auditWriter *audit.Writer) *PermissionService {
	return &PermissionService{q: q, pool: pool, audit: auditWriter}
}

// parsePermission parses "resource:action" (e.g. "order:read", "order:*") into resource and action.
func parsePermission(s string) (resource, action string, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", "", ErrPermissionInvalidFormat
	}
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", ErrPermissionInvalidFormat
	}
	return parts[0], parts[1], nil
}

// PermissionToString returns "resource:action" for a permission.
func PermissionToString(resource, action string) string {
	return resource + ":" + action
}

// ReplaceRolePermissions replaces all permissions for a role. Returns ErrRoleNotFound if role does not exist.
func (s *PermissionService) ReplaceRolePermissions(
	ctx context.Context,
	tenantID, roleID string,
	permissions []string,
	actorID string,
) ([]string, error) {
	tid, err := parseUUID(tenantID)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	rid, err := parseUUID(roleID)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	_, err = s.q.GetRole(ctx, db.GetRoleParams{ID: rid, TenantID: tid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	// Parse and validate all permissions first
	var parsed [][2]string
	for _, p := range permissions {
		var res, act string
		res, act, err = parsePermission(p)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, [2]string{res, act})
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			slog.Debug("tx rollback", "error", rbErr)
		}
	}()
	q := db.New(tx)
	if err := q.DeleteRolePermissions(ctx, rid); err != nil {
		return nil, err
	}
	var result []string
	for _, p := range parsed {
		perm, err := q.CreatePermission(ctx, db.CreatePermissionParams{
			TenantID: tid,
			Resource: p[0],
			Action:   p[1],
		})
		if err != nil {
			return nil, err
		}
		if err := q.AddRolePermission(ctx, db.AddRolePermissionParams{
			RoleID:       rid,
			PermissionID: perm.ID,
		}); err != nil {
			return nil, err
		}
		result = append(result, PermissionToString(perm.Resource, perm.Action))
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	if s.audit != nil && actorID != "" {
		summary := fmt.Sprintf("replaced with %d permission(s)", len(result))
		if err := s.audit.AppendAdminEvent(ctx, actorID, "role.permissions.replace", "role", roleID, summary, tid); err != nil {
			slog.Warn("audit write failed", "action", "role.permissions.replace", "error", err)
		}
	}
	return result, nil
}

// AddRolePermissions adds permissions to a role. Returns ErrRoleNotFound if role does not exist.
func (s *PermissionService) AddRolePermissions(
	ctx context.Context,
	tenantID, roleID string,
	permissions []string,
	actorID string,
) ([]string, error) {
	tid, err := parseUUID(tenantID)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	rid, err := parseUUID(roleID)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	_, err = s.q.GetRole(ctx, db.GetRoleParams{ID: rid, TenantID: tid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	var result []string
	for _, p := range permissions {
		res, act, parseErr := parsePermission(p)
		if parseErr != nil {
			return nil, parseErr
		}
		perm, err := s.q.CreatePermission(ctx, db.CreatePermissionParams{
			TenantID: tid,
			Resource: res,
			Action:   act,
		})
		if err != nil {
			return nil, err
		}
		if err := s.q.AddRolePermission(ctx, db.AddRolePermissionParams{
			RoleID:       rid,
			PermissionID: perm.ID,
		}); err != nil {
			return nil, err
		}
		result = append(result, PermissionToString(perm.Resource, perm.Action))
	}
	if s.audit != nil && actorID != "" && len(result) > 0 {
		summary := fmt.Sprintf("added %d permission(s): %s", len(result), strings.Join(result, ", "))
		if err := s.audit.AppendAdminEvent(ctx, actorID, "role.permissions.add", "role", roleID, summary, tid); err != nil {
			slog.Warn("audit write failed", "action", "role.permissions.add", "error", err)
		}
	}
	return result, nil
}

// RemoveRolePermissions removes permissions from a role. Returns ErrRoleNotFound if role does not exist.
func (s *PermissionService) RemoveRolePermissions(
	ctx context.Context,
	tenantID, roleID string,
	permissions []string,
	actorID string,
) ([]string, error) {
	tid, err := parseUUID(tenantID)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	rid, err := parseUUID(roleID)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	_, err = s.q.GetRole(ctx, db.GetRoleParams{ID: rid, TenantID: tid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	var removed []string
	for _, p := range permissions {
		res, act, parseErr := parsePermission(p)
		if parseErr != nil {
			return nil, parseErr
		}
		perm, err := s.q.GetPermissionByTenantResourceAction(ctx, db.GetPermissionByTenantResourceActionParams{
			TenantID: tid,
			Resource: res,
			Action:   act,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return nil, err
		}
		if err := s.q.RemoveRolePermission(ctx, db.RemoveRolePermissionParams{
			RoleID:       rid,
			PermissionID: perm.ID,
		}); err != nil {
			return nil, err
		}
		removed = append(removed, PermissionToString(perm.Resource, perm.Action))
	}
	if s.audit != nil && actorID != "" && len(removed) > 0 {
		summary := fmt.Sprintf("removed %d permission(s): %s", len(removed), strings.Join(removed, ", "))
		if err := s.audit.AppendAdminEvent(ctx, actorID, "role.permissions.remove", "role", roleID, summary, tid); err != nil {
			slog.Warn("audit write failed", "action", "role.permissions.remove", "error", err)
		}
	}
	return removed, nil
}

// ListRolePermissions returns the list of permissions for a role as "resource:action" strings.
func (s *PermissionService) ListRolePermissions(ctx context.Context, tenantID, roleID string) ([]string, error) {
	tid, err := parseUUID(tenantID)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	rid, err := parseUUID(roleID)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	_, err = s.q.GetRole(ctx, db.GetRoleParams{ID: rid, TenantID: tid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	perms, err := s.q.ListPermissionsByRole(ctx, db.ListPermissionsByRoleParams{
		RoleID:   rid,
		TenantID: tid,
	})
	if err != nil {
		return nil, err
	}
	result := make([]string, len(perms))
	for i, p := range perms {
		result[i] = PermissionToString(p.Resource, p.Action)
	}
	return result, nil
}
