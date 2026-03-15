-- Roles CRUD queries
-- Source: db-schema.md §2.4

-- name: CreateRole :one
INSERT INTO roles (tenant_id, name, description, status, is_builtin)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, name, description, status, is_builtin, created_at, updated_at;

-- name: GetRole :one
SELECT id, tenant_id, name, description, status, is_builtin, created_at, updated_at
FROM roles
WHERE id = $1 AND tenant_id = $2;

-- name: ListRolesByTenant :many
-- Cursor: pass uuid.Nil or zero UUID for first page
SELECT id, tenant_id, name, description, status, is_builtin, created_at, updated_at
FROM roles
WHERE tenant_id = $1 AND id > $2
ORDER BY id
LIMIT $3;

-- name: UpdateRole :one
UPDATE roles
SET name = $2, description = $3, status = $4, updated_at = now()
WHERE id = $1 AND tenant_id = $5
RETURNING id, tenant_id, name, description, status, is_builtin, created_at, updated_at;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1 AND tenant_id = $2;

-- name: GetRoleByNameAndTenant :one
SELECT id, tenant_id, name, description, status, is_builtin, created_at, updated_at
FROM roles
WHERE tenant_id = $1 AND name = $2;
