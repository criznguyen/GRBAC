# ADR-004: Audit Storage and Retention

## Status
Accepted

## Context

- **FR-016:** Record audit log for every permission check (subject, resource, action, timestamp, result, tenant).
- **FR-017:** Record audit log for every RBAC admin action (who, what, when).
- **FR-018:** API to export audit log (CSV/JSON) with filters and pagination.
- **NFR:** Audit retained at least 1 year (configurable); support GDPR, PCI-DSS, SOC2 (immutable log, query/export).

We need to decide: where audit is stored (already decided in ADR-002: separate PostgreSQL), how to write it without impacting p99, retention policy, and indexing for query/export.

## Decision

1. **Storage:** Audit events are stored in the **Audit Store** (PostgreSQL, separate from Policy Store) as in ADR-002. Tables (or logical streams): one for **permission check events**, one for **admin action events**. Schema includes: tenant_id, subject_id, resource, action, decision (Allow/Deny), timestamp, request_id, scope (optional) for checks; actor_id, action_type, target_type, target_id, change_summary (or before/after), timestamp, tenant_id for admin events.
2. **Write path and p99:** **Asynchronous write** for permission check audit: PDP returns Allow/Deny to the client first, then enqueues the audit record (in-process queue or message queue). A separate writer process (or same process with background flush) appends to Audit Store. This keeps audit write off the critical path so p99 < 50ms is not dominated by DB write. Admin action audit can be **synchronous** (after DB commit) because admin operations are not latency-critical; alternatively async for consistency. Technical BA to specify: async queue (e.g. in-memory + durable queue or Kafka) and at-least-once delivery.
3. **Retention:** Configurable retention period; **minimum 1 year** default. Implemented by partitioning audit tables by time (e.g. month). Retention job drops or archives partitions older than retention period. Export API respects retention (no export of purged data).
4. **Immutability:** No UPDATE or DELETE on audit rows. Access control so only the audit writer can INSERT; query/export is read-only. Backups and replication support compliance and recovery.
5. **Query and export:** Indexes on (tenant_id, event_time), (subject_id, event_time), (resource, action, event_time), (actor_id, event_time) to support filter combinations. Export uses same indexes with pagination (limit/offset or cursor); large exports can be async (job id + download link) as per FR-018.

## Consequences

- **Positive:** p99 not impacted by audit write; immutable append-only supports compliance; partitioning and retention keep storage bounded; query/export performance via indexes.
- **Negative:** Async write implies eventual consistency for audit (small delay before event appears in query); need to handle queue backlog and failure (retry, dead-letter). Technical BA must define durability and at-least-once semantics.
- **Follow-up:** Technical BA to specify audit schema, async write mechanism, retention job, and export API (sync vs async, limits).
