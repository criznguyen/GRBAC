# Database Schema — RBAC for Microservices

**Version:** 1.0  
**Date:** 2025-03-15  
**Author:** Technical BA  
**Source:** HANDOFF-TO-TECHNICAL-BA.md, ADR-002, ADR-004, ADR-005  
**Status:** Ready for Dev

---

## 1. Overview

Two PostgreSQL databases:

1. **Policy Store:** Roles, permissions, users, groups, assignments, delegation, resource registry. Source of truth for policy.
2. **Audit Store:** Append-only audit events (checks + admin). Partitioned by time; retention configurable.

All tables include `tenant_id` for multi-tenancy (ADR-005).

---

## 2. Policy Store Schema

### 2.1 Tenants

```sql
CREATE TABLE tenants (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name       VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 2.2 Users

```sql
CREATE TABLE users (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  external_id VARCHAR(255),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, external_id)
);

CREATE INDEX idx_users_tenant ON users(tenant_id);
```

### 2.3 Groups

```sql
CREATE TABLE groups (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name        VARCHAR(255) NOT NULL,
  description TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, name)
);

CREATE INDEX idx_groups_tenant ON groups(tenant_id);
```

### 2.4 Roles

```sql
CREATE TABLE roles (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name        VARCHAR(255) NOT NULL,
  description TEXT,
  status      VARCHAR(50) NOT NULL DEFAULT 'active',
  is_builtin  BOOLEAN NOT NULL DEFAULT false,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, name)
);

CREATE INDEX idx_roles_tenant ON roles(tenant_id);
CREATE INDEX idx_roles_tenant_status ON roles(tenant_id, status);
```

### 2.5 Permissions (canonical resource:action)

```sql
CREATE TABLE permissions (
  id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  resource  VARCHAR(255) NOT NULL,
  action    VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, resource, action)
);

CREATE INDEX idx_permissions_tenant ON permissions(tenant_id);
```

### 2.6 Role Permissions

```sql
CREATE TABLE role_permissions (
  role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);
```

### 2.7 Role Hierarchy (parent-child)

```sql
CREATE TABLE role_hierarchy (
  child_role_id  UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  parent_role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  PRIMARY KEY (child_role_id),
  CHECK (child_role_id != parent_role_id)
);

CREATE INDEX idx_role_hierarchy_parent ON role_hierarchy(parent_role_id);
```

### 2.8 User Roles

```sql
CREATE TABLE user_roles (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id     UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  scope_type  VARCHAR(50),
  scope_id    VARCHAR(255),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, user_id, role_id, scope_type, scope_id)
);

CREATE INDEX idx_user_roles_tenant_user ON user_roles(tenant_id, user_id);
CREATE INDEX idx_user_roles_tenant_role ON user_roles(tenant_id, role_id);
CREATE INDEX idx_user_roles_scope ON user_roles(tenant_id, user_id, scope_type, scope_id);
```

### 2.9 Group Roles

```sql
CREATE TABLE group_roles (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  group_id    UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  role_id     UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  scope_type  VARCHAR(50),
  scope_id    VARCHAR(255),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, group_id, role_id, scope_type, scope_id)
);

CREATE INDEX idx_group_roles_tenant_group ON group_roles(tenant_id, group_id);
CREATE INDEX idx_group_roles_tenant_role ON group_roles(tenant_id, role_id);
```

### 2.10 Group Members

```sql
CREATE TABLE group_members (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  group_id   UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, group_id, user_id)
);

CREATE INDEX idx_group_members_tenant_group ON group_members(tenant_id, group_id);
CREATE INDEX idx_group_members_tenant_user ON group_members(tenant_id, user_id);
```

### 2.11 Delegations

```sql
CREATE TABLE delegations (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id          UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  delegator_user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id            UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  scope_type         VARCHAR(50) NOT NULL,
  scope_id           VARCHAR(255) NOT NULL,
  allow_assign       BOOLEAN NOT NULL DEFAULT true,
  allow_revoke       BOOLEAN NOT NULL DEFAULT true,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_delegations_tenant_delegator ON delegations(tenant_id, delegator_user_id);
CREATE INDEX idx_delegations_tenant_role ON delegations(tenant_id, role_id);
CREATE INDEX idx_delegations_scope ON delegations(tenant_id, scope_type, scope_id);
```

### 2.12 Resource Registry

```sql
CREATE TABLE resource_registry (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  resource_type VARCHAR(255) NOT NULL,
  actions      JSONB NOT NULL,
  description  TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, resource_type)
);

CREATE INDEX idx_resource_registry_tenant ON resource_registry(tenant_id);
```

---

## 3. Audit Store Schema

### 3.1 Permission Check Events (Partitioned)

```sql
CREATE TABLE audit_checks (
  id         UUID NOT NULL DEFAULT gen_random_uuid(),
  tenant_id  UUID NOT NULL,
  subject_id VARCHAR(255) NOT NULL,
  resource   VARCHAR(255) NOT NULL,
  action     VARCHAR(255) NOT NULL,
  decision   VARCHAR(10) NOT NULL CHECK (decision IN ('Allow', 'Deny')),
  scope_type VARCHAR(50),
  scope_id   VARCHAR(255),
  event_time TIMESTAMPTZ NOT NULL DEFAULT now(),
  request_id VARCHAR(255)
) PARTITION BY RANGE (event_time);

-- Example partition (create monthly)
CREATE TABLE audit_checks_2025_03 PARTITION OF audit_checks
  FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

CREATE INDEX idx_audit_checks_tenant_time ON audit_checks(tenant_id, event_time);
CREATE INDEX idx_audit_checks_subject_time ON audit_checks(subject_id, event_time);
CREATE INDEX idx_audit_checks_resource_action_time ON audit_checks(resource, action, event_time);
```

### 3.2 Admin Action Events (Partitioned)

```sql
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

CREATE TABLE audit_admin_2025_03 PARTITION OF audit_admin
  FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

CREATE INDEX idx_audit_admin_tenant_time ON audit_admin(tenant_id, event_time);
CREATE INDEX idx_audit_admin_actor_time ON audit_admin(actor_id, event_time);
CREATE INDEX idx_audit_admin_action_type_time ON audit_admin(action_type, event_time);
```

---

## 4. Partitioning & Retention

- **Partition key:** `event_time` (monthly partitions recommended).
- **Retention job:** Drop partitions older than retention period (configurable, min 1 year per ADR-004).
- **Naming:** `audit_checks_YYYY_MM`, `audit_admin_YYYY_MM`.

---

## 5. Index Summary

| Table | Index | Purpose |
|-------|-------|---------|
| users | (tenant_id) | List users, tenant lookup |
| groups | (tenant_id) | List groups |
| roles | (tenant_id), (tenant_id, status) | List roles, filter active |
| user_roles | (tenant_id, user_id), (tenant_id, role_id), (tenant_id, user_id, scope_type, scope_id) | Subject lookup, revoke by role, scope filter |
| group_roles | (tenant_id, group_id), (tenant_id, role_id) | Group roles, role revocation |
| group_members | (tenant_id, group_id), (tenant_id, user_id) | Resolve user→groups, group membership |
| delegations | (tenant_id, delegator_user_id), (tenant_id, scope_type, scope_id) | Delegation check |
| audit_checks | (tenant_id, event_time), (subject_id, event_time), (resource, action, event_time) | Query, export filters |
| audit_admin | (tenant_id, event_time), (actor_id, event_time), (action_type, event_time) | Query, export filters |

---

## 6. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | Technical BA | Initial DB schema |
