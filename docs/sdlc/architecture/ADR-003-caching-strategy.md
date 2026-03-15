# ADR-003: Caching Strategy (Redis, In-Memory, TTL)

## Status
Accepted

## Context

- **FR-023:** Cache layer for permission decisions and/or policy data; configurable TTL; cache invalidated when policy or assignments change so results stay correct.
- **FR-024:** PDP scales horizontally; no per-instance state for decision logic.
- **NFR:** p99 < 50ms for permission check.

We need to decide: cache backend (Redis vs in-memory only), what to cache (decisions vs policy data), cache key shape, TTL, and invalidation strategy.

## Decision

1. **Backend:** **Redis** as the shared cache for the central PDP. All PDP instances use the same Redis (or Redis cluster) so that (a) cache is shared across instances, (b) invalidation from Admin API is globally visible. SDK may additionally use a **local in-memory cache** (per process) with shorter TTL as in ADR-001.
2. **What to cache:** **Permission check results** (Allow/Deny) keyed by (tenant_id, subject_id, resource, action, scope). Optionally cache **resolved permission set per subject** (tenant_id, subject_id, scope) to speed up cache misses when the same subject checks multiple resources; this is an implementation detail for Technical BA.
3. **Cache key shape:** `rbac:check:{tenant_id}:{subject_id}:{resource}:{action}:{scope_hash}` for decision cache. Scope can be empty (global). Subject_id is the resolved user id (and scope distinguishes org/project if used).
4. **TTL:** Configurable; default **60 seconds** for decision cache. Short enough to limit staleness after invalidation misses; long enough to reduce load and meet p99. Technical BA to make TTL configurable per deployment.
5. **Invalidation:** On every policy/assignment change (role CRUD, assign/revoke permission, assign/revoke role to user/group, group membership, delegation, policy import), Admin API triggers invalidation:
   - **By subject:** Invalidate all keys for affected subject_id(s) (e.g. user or users in a group).
   - **By role:** Invalidate all keys for subjects that had that role (pattern delete or maintain role→subjects index for targeted invalidation).
   - **By tenant:** If needed, invalidate all keys for tenant (e.g. bulk policy import).

   Prefer **targeted invalidation** (by subject/role) over full-tenant flush to avoid thundering herd. Use Redis pattern delete or a small set of “invalidation version” keys per subject that invalidate logical cache entries (Technical BA to specify exact mechanism).

## Consequences

- **Positive:** Shared Redis gives consistent view across PDP instances; invalidation on change keeps results correct; TTL provides safety net; p99 target achievable with cache hits.
- **Negative:** Redis is an additional dependency and single point of failure unless run as cluster/sentinel; invalidation logic must be correct and tested (e.g. group membership change invalidates all members).
- **Follow-up:** Technical BA to specify Redis key namespace, invalidation API or events, and optional “per-subject permission set” cache design; document TTL and invalidation in API/spec.
