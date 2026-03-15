//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// AuditEvent represents a row from audit_admin for assertions.
type AuditEvent struct {
	ActionType string
	TargetType string
	TargetID   string
	ActorID    string
	TenantID   string
}

// QueryAuditAdminEvents returns audit_admin events for the given tenant, ordered by event_time desc.
func QueryAuditAdminEvents(t *testing.T, dbURL string, tenantID string) []AuditEvent {
	t.Helper()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	parsed, err := uuid.Parse(tenantID)
	require.NoError(t, err)

	rows, err := pool.Query(ctx,
		`SELECT action_type, target_type, target_id, actor_id, tenant_id
		 FROM audit_admin
		 WHERE tenant_id = $1
		 ORDER BY event_time DESC`,
		parsed,
	)
	require.NoError(t, err)
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var e AuditEvent
		var targetType, targetID *string
		var actorID string
		var tid uuid.UUID
		err := rows.Scan(&e.ActionType, &targetType, &targetID, &actorID, &tid)
		require.NoError(t, err)
		e.ActorID = actorID
		e.TenantID = tid.String()
		if targetType != nil {
			e.TargetType = *targetType
		}
		if targetID != nil {
			e.TargetID = *targetID
		}
		events = append(events, e)
	}
	require.NoError(t, rows.Err())
	return events
}
