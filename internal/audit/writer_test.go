package audit

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grbac/grbac/internal/db"
)

// mockInserter implements Inserter for testing.
type mockInserter struct {
	insertFunc func(ctx context.Context, arg db.InsertAuditAdminParams) (db.AuditAdmin, error)
}

func (m *mockInserter) InsertAuditAdmin(ctx context.Context, arg db.InsertAuditAdminParams) (db.AuditAdmin, error) {
	if m.insertFunc != nil {
		return m.insertFunc(ctx, arg)
	}
	return db.AuditAdmin{}, nil
}

func TestAppendAdminEvent_WritesEvent(t *testing.T) {
	tid := uuid.New()
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	var captured db.InsertAuditAdminParams
	ins := &mockInserter{
		insertFunc: func(ctx context.Context, arg db.InsertAuditAdminParams) (db.AuditAdmin, error) {
			captured = arg
			return db.AuditAdmin{}, nil
		},
	}
	w := NewWriter(ins)

	err := w.AppendAdminEvent(
		context.Background(),
		"actor-123",
		"role.create",
		"role",
		"role-uuid",
		"created role \"Order Manager\"",
		pgTid,
	)
	require.NoError(t, err)

	assert.Equal(t, pgTid, captured.TenantID)
	assert.Equal(t, "actor-123", captured.ActorID)
	assert.Equal(t, "role.create", captured.ActionType)
	require.True(t, captured.TargetType.Valid)
	assert.Equal(t, "role", captured.TargetType.String)
	require.True(t, captured.TargetID.Valid)
	assert.Equal(t, "role-uuid", captured.TargetID.String)
	require.True(t, captured.ChangeSummary.Valid)
	assert.Equal(t, "created role \"Order Manager\"", captured.ChangeSummary.String)
}

func TestAppendAdminEvent_HandlesEmptyOptionalFields(t *testing.T) {
	tid := uuid.New()
	var pgTid pgtype.UUID
	pgTid.Bytes = tid
	pgTid.Valid = true

	var captured db.InsertAuditAdminParams
	ins := &mockInserter{
		insertFunc: func(ctx context.Context, arg db.InsertAuditAdminParams) (db.AuditAdmin, error) {
			captured = arg
			return db.AuditAdmin{}, nil
		},
	}
	w := NewWriter(ins)

	err := w.AppendAdminEvent(
		context.Background(),
		"actor-456",
		"role.permissions.replace",
		"",
		"",
		"",
		pgTid,
	)
	require.NoError(t, err)

	assert.Equal(t, "actor-456", captured.ActorID)
	assert.Equal(t, "role.permissions.replace", captured.ActionType)
	assert.False(t, captured.TargetType.Valid)
	assert.False(t, captured.TargetID.Valid)
	assert.False(t, captured.ChangeSummary.Valid)
}
