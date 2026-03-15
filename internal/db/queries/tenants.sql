-- Tenants CRUD queries
-- Source: db-schema.md §2.1

-- name: GetTenant :one
SELECT id, name, created_at
FROM tenants
WHERE id = $1;

-- name: CreateTenant :one
INSERT INTO tenants (name)
VALUES ($1)
RETURNING id, name, created_at;
