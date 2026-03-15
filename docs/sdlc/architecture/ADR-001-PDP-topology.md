# ADR-001: PDP Topology (Centralized vs Sidecar vs Hybrid)

## Status
Accepted

## Context

The RBAC service must expose a Policy Decision Point (PDP) so microservices can get Allow/Deny decisions for (subject, resource, action). Requirements:

- **FR-011:** Check Permission API; batch check supported.
- **FR-012:** SDKs (Node, Go, Java) with simple `check(userId, resource, action)`.
- **FR-013:** PDP as standalone component, separate from policy administration; scale independently; policy updates reflected in real time without redeploying consuming services.
- **NFR:** p99 < 50ms for permission check; 99.9% availability.

We need to choose where the PDP runs relative to the calling service: **centralized** (all checks go to a shared RBAC service), **sidecar** (PDP runs next to each service), or **hybrid** (central PDP with optional edge/sidecar cache).

## Decision

We adopt a **hybrid topology**:

1. **Primary:** **Centralized PDP** — a dedicated RBAC service (or cluster) that exposes the Check Permission API (REST and/or gRPC). All microservices (PEPs) call this central PDP. The PDP is stateless and scales horizontally behind a load balancer.
2. **SDK:** SDKs call the central PDP over the network. SDKs may optionally use a **local in-memory cache** (e.g. per process) with short TTL to reduce latency and load; cache key: tenant+subject+resource+action+scope. Cache invalidation is best-effort via TTL; critical path does not depend on push invalidation to the SDK.
3. **No sidecar PDP in Phase 1:** We do not deploy a separate PDP process per service (e.g. Envoy sidecar with embedded PDP). This keeps operations simple and avoids per-pod policy sync. Sidecar can be revisited in a later phase if latency or network isolation demands it.

Rationale:

- **Centralized PDP** gives a single place to apply policy updates (no pushing policy to N sidecars), meets FR-013 (standalone, scalable), and simplifies audit (all decisions logged at the service).
- **Hybrid with SDK-level cache** improves p99 for repeated checks and reduces load on the central PDP while keeping implementation and ops manageable.
- **No sidecar in Phase 1** avoids operational complexity and aligns with “integrate in < 1 day” and 99.9% SLA using a well-known pattern (central service + LB + cache).

## Consequences

- **Positive:** Single source of policy; straightforward deployment and scaling; SDK cache improves latency for hot keys; audit remains centralized; no per-pod policy distribution.
- **Negative:** Every check (on cache miss) has network hop to central PDP; SDK cache can be stale up to TTL (mitigated by short TTL and invalidation scope in ADR-003).
- **Follow-up:** Technical BA to specify Check Permission API contract (REST + gRPC), SDK cache TTL and key shape, and retry/timeout behavior.
