// Package audit provides the admin audit writer (sync for MVP).
// Writes to audit_admin table. Work package (3).
package audit

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/grbac/grbac/internal/db"
)

// Inserter defines the minimal interface for writing audit events.
// Implemented by db.Queries; allows mocking in tests.
type Inserter interface {
	InsertAuditAdmin(ctx context.Context, arg db.InsertAuditAdminParams) (db.AuditAdmin, error)
}

// Writer appends admin action events to audit_admin.
type Writer struct {
	ins Inserter
}

// NewWriter creates an audit writer that writes to the given inserter.
// For MVP, pass db.Queries (same DB as policy store). For separate audit DB, pass Queries from audit pool.
func NewWriter(ins Inserter) *Writer {
	return &Writer{ins: ins}
}

// AppendAdminEvent synchronously writes an admin action event to audit_admin.
// actorID is the subject from auth (e.g. JWT sub). targetType/targetID identify the affected entity.
func (w *Writer) AppendAdminEvent(
	ctx context.Context,
	actorID string,
	actionType string,
	targetType string,
	targetID string,
	changeSummary string,
	tenantID pgtype.UUID,
) error {
	targetTypeVal := pgtype.Text{}
	if targetType != "" {
		targetTypeVal.String = targetType
		targetTypeVal.Valid = true
	}
	targetIDVal := pgtype.Text{}
	if targetID != "" {
		targetIDVal.String = targetID
		targetIDVal.Valid = true
	}
	changeSummaryVal := pgtype.Text{}
	if changeSummary != "" {
		changeSummaryVal.String = changeSummary
		changeSummaryVal.Valid = true
	}
	_, err := w.ins.InsertAuditAdmin(ctx, db.InsertAuditAdminParams{
		TenantID:      tenantID,
		ActorID:       actorID,
		ActionType:    actionType,
		TargetType:    targetTypeVal,
		TargetID:      targetIDVal,
		ChangeSummary: changeSummaryVal,
	})
	return err
}
