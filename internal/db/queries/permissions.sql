-- Permissions and role_permissions queries
-- Source: db-schema.md §§2.5, 2.6

-- name: CreatePermission :one
INSERT INTO permissions (tenant_id, resource, action)
VALUES ($1, $2, $3)
ON CONFLICT (tenant_id, resource, action) DO UPDATE SET resource = EXCLUDED.resource
RETURNING id, tenant_id, resource, action, created_at;

-- name: GetPermissionByIDs :one
SELECT id, tenant_id, resource, action, created_at
FROM permissions
WHERE id = $1 AND tenant_id = $2;

-- name: GetPermissionByTenantResourceAction :one
SELECT id, tenant_id, resource, action, created_at
FROM permissions
WHERE tenant_id = $1 AND resource = $2 AND action = $3;

-- name: ListPermissionsByRole :many
SELECT p.id, p.tenant_id, p.resource, p.action, p.created_at
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
WHERE rp.role_id = $1 AND p.tenant_id = $2;

-- name: DeleteRolePermissions :exec
DELETE FROM role_permissions WHERE role_id = $1;

-- name: AddRolePermission :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- name: RemoveRolePermission :exec
DELETE FROM role_permissions
WHERE role_id = $1 AND permission_id = $2;
