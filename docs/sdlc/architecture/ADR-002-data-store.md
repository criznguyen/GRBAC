# ADR-002: Data Store Choice (Policy and Audit)

## Status
Accepted

## Context

The RBAC system needs persistent storage for:

1. **Policy data:** Roles, permissions, role-permission assignments, user-role and group-role assignments, groups, membership, delegation rules, role hierarchy, resource registry. Requires CRUD, transactional updates, and strong consistency for admin operations.
2. **Audit data:** Append-only log of (a) every permission check (subject, resource, action, decision, timestamp, tenant), (b) every admin action (actor, action type, target, before/after, timestamp). Must support query and export by filters (date, user, resource, tenant); retention at least 1 year (configurable); compliance (GDPR, PCI-DSS, SOC2).

NFRs: 99.9% availability; audit immutable and queryable; multi-tenant isolation.

## Decision

- **Policy Store:** **PostgreSQL**. Single primary database (with replication for HA) holding all policy entities. All tables are keyed by `tenant_id` for multi-tenancy (row-level isolation per ADR-005). Use standard ACID transactions for admin operations; PDP reads (on cache miss) are read-only and can use replicas if needed.
- **Audit Store:** **PostgreSQL** in a **separate database** (or separate schema with dedicated retention/backup). Append-only tables; no updates or deletes on audit rows. Partitioning by time (e.g. month) for efficient query and retention purge. Indexes on (tenant_id, timestamp), (subject_id, timestamp), (resource, action, timestamp) to support audit query and export filters.

Rationale:

- **PostgreSQL** provides ACID, JSON support (for flexible audit payloads), partitioning, and mature replication/backup. Team familiarity and operational maturity support 99.9% SLA.
- **Separate audit database** isolates audit from policy workload, allows different backup/retention and compliance handling, and avoids lock contention between heavy admin/check traffic and audit writes.
- **Append-only + partitioning** keeps audit immutable and enables efficient range queries and retention (drop old partitions or archive).

## Consequences

- **Positive:** Strong consistency for policy; clear separation of policy vs audit; audit lifecycle (retention, export) can be managed independently; good fit for compliance requirements.
- **Negative:** Two PostgreSQL clusters to operate (or two databases in one cluster); Technical BA must define schema and indexing for both.
- **Follow-up:** Technical BA to specify schema (tables, indexes, partitioning), connection handling, and read-replica usage for PDP reads if applicable.
