-- Add audit partition for 2026 (required for tests and production in 2026)
CREATE TABLE IF NOT EXISTS audit_admin_2026_03 PARTITION OF audit_admin
  FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
