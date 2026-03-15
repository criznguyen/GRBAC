# ADR-005: Multi-Tenancy Isolation Strategy

## Status
Accepted

## Context

- **FR-014:** Multi-tenancy; all entities (users, roles, permissions, groups) scoped to tenant; API includes tenant context (e.g. X-Tenant-ID).
- **FR-015:** Optional scope (org, project) for role assignments and permission checks.
- **NFR:** No cross-tenant data leakage; compliance (e.g. GDPR) per tenant.

We need to choose how to isolate tenant data: single database with tenant_id column (row-level), tenant-scoped schemas, or database-per-tenant.

## Decision

We use **single database (Policy Store) with row-level isolation by tenant_id**:

1. **Policy Store:** One PostgreSQL database. Every policy table includes **tenant_id** (NOT NULL) as a leading column. All queries (Admin API and PDP) **always** filter by tenant_id derived from request context (header, token claim, or routing). No query ever returns or updates rows from another tenant. Indexes are tenant-first (e.g. (tenant_id, role_id)) for performance.
2. **Audit Store:** Same approach: **tenant_id** on every audit row; all query and export operations scoped by tenant_id. Retention and partitioning can still be time-based; tenant_id is used for filtering.
3. **No schema-per-tenant or database-per-tenant in Phase 1:** We avoid separate schemas or databases per tenant to keep operations and schema migrations simple. Single schema with tenant_id is sufficient for isolation when enforced at application and API layer; RBAC does not expose tenant_id enumeration or cross-tenant queries.
4. **Scope (org/project):** Scope is an **optional** dimension on assignments and on check requests. Stores as columns (e.g. scope_type, scope_id) or JSON where needed. Permission check filters assignments by scope when scope is provided; otherwise treats as “global” within tenant. No separate isolation boundary—scope is logical filter within a tenant.

Security and validation:

- **Tenant context mandatory:** Every API request must carry tenant_id (validated against tenant registry or token). Reject requests with missing or invalid tenant.
- **Authorization:** Admin API must ensure caller is allowed to act in that tenant (RBAC or equivalent); PDP only returns Allow/Deny for the tenant in the request.
- **Audit:** Every audit row includes tenant_id; export and query are tenant-scoped for the caller’s tenant(s).

## Consequences

- **Positive:** Simple operations (one policy DB, one audit DB); straightforward migrations; tenant_id indexing keeps queries efficient; well-understood pattern.
- **Negative:** No hard physical isolation between tenants (mitigated by strict app-level filtering and auth); very large tenants could theoretically impact others (monitoring and indexing help).
- **Follow-up:** Technical BA to define tenant_id presence in every entity and API contract; validation and error handling for missing/invalid tenant; scope model (columns, indexing) for FR-015.
