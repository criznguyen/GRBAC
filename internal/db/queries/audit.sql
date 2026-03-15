-- Audit admin events
-- Source: db-schema.md §3.2

-- name: InsertAuditAdmin :one
INSERT INTO audit_admin (tenant_id, actor_id, action_type, target_type, target_id, change_summary)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, actor_id, action_type, target_type, target_id, change_summary, event_time, metadata;
