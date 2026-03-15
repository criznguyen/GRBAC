-- Audit Store — admin action events (partitioned by event_time)
-- Source: db-schema.md §3.2

CREATE TABLE audit_admin (
  id             UUID NOT NULL DEFAULT gen_random_uuid(),
  tenant_id      UUID NOT NULL,
  actor_id       VARCHAR(255) NOT NULL,
  action_type    VARCHAR(100) NOT NULL,
  target_type    VARCHAR(100),
  target_id      VARCHAR(255),
  change_summary TEXT,
  event_time     TIMESTAMPTZ NOT NULL DEFAULT now(),
  metadata       JSONB
) PARTITION BY RANGE (event_time);

-- Initial partition: 2025-03
CREATE TABLE audit_admin_2025_03 PARTITION OF audit_admin
  FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

CREATE INDEX idx_audit_admin_tenant_time ON audit_admin(tenant_id, event_time);
CREATE INDEX idx_audit_admin_actor_time ON audit_admin(actor_id, event_time);
CREATE INDEX idx_audit_admin_action_type_time ON audit_admin(action_type, event_time);
