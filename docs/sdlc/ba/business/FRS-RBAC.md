# Functional Requirements Specification (FRS)
# RBAC for Microservices

**Version:** 1.0  
**Date:** 2025-03-15  
**Author:** Business BA  
**Source:** PRD-RBAC-Microservices.md, epic-brief-RBAC.md  
**Status:** Ready for Architect

---

## 1. Introduction

This document breaks down the Product Owner's Epics and User Stories into detailed Functional Requirements (FRs). Each FR is traceable to the PRD and includes Description, Trigger, Process Flow, Output, and Constraints.

---

## 2. Traceability Summary

| Epic | PRD Reference | FRs |
|------|---------------|-----|
| Epic 1: Core RBAC Model | §6 Epic 1 | FR-001 – FR-007 |
| Epic 2: Groups & Delegation | §6 Epic 2 | FR-008 – FR-010 |
| Epic 3: Policy Decision & Enforcement | §6 Epic 3 | FR-011 – FR-013 |
| Epic 4: Multi-Tenancy & Scope | §6 Epic 4 | FR-014 – FR-015 |
| Epic 5: Audit & Compliance | §6 Epic 5 | FR-016 – FR-018 |
| Epic 6: Resource Registry & Policy as Code | §6 Epic 6 | FR-019 – FR-020 |
| Epic 7: Built-in Roles & Default Policies | §6 Epic 7 | FR-021 – FR-022 |
| Epic 8: Performance & Scalability | §6 Epic 8 | FR-023 – FR-024 |

---

## 3. Functional Requirements

### Epic 1: Core RBAC Model

---

#### FR-001: Create Role

**Traceability:** Epic 1, E1-US1

**Description:** The system shall allow a System Admin to create a new Role with a unique name and description within a tenant, so that permissions can be grouped by job function.

**Trigger:** System Admin submits a request to create a role (name, description, optional status).

**Process Flow:**
1. Admin invokes Create Role API with tenant context (e.g. `X-Tenant-ID`).
2. System validates that role name is unique within the tenant.
3. System creates the role with name, description, status (e.g. active).
4. System records audit event (admin action).
5. System returns the created role (id, name, description, status).

**Output:** Created role resource (id, name, description, status); 201 on success.

**Constraints:** Role name must be unique per tenant; name and description required; compliance with audit trail (E5-US2).

---

#### FR-002: Update Role

**Traceability:** Epic 1, E1-US1

**Description:** The system shall allow a System Admin to update an existing Role's name, description, or status.

**Trigger:** System Admin submits an update request for a role by id.

**Process Flow:**
1. Admin invokes Update Role API with role id and new attributes.
2. System validates role exists and name (if changed) remains unique per tenant.
3. System updates the role.
4. System records audit event.
5. System returns the updated role.

**Output:** Updated role resource; 200 on success.

**Constraints:** Unique name per tenant; audit trail required.

---

#### FR-003: Delete Role

**Traceability:** Epic 1, E1-US1

**Description:** The system shall allow a System Admin to delete a Role, with consistent handling of existing assignments.

**Trigger:** System Admin submits delete request for a role by id.

**Process Flow:**
1. Admin invokes Delete Role API with role id.
2. System validates role exists and optionally checks for existing user/group assignments.
3. System removes role and any role-permission and role-assignment links (or marks inactive per policy).
4. System records audit event.
5. System returns success.

**Output:** 204 No Content or 200 with confirmation.

**Constraints:** Policy may require revoking assignments first or cascade delete; audit trail required.

---

#### FR-004: Assign Permissions to Role

**Traceability:** Epic 1, E1-US2

**Description:** The system shall allow a System Admin to assign Permissions to a Role. A permission is a pair (resource, action), e.g. `order:read`, `user:delete`, with optional wildcard support (`*`).

**Trigger:** System Admin requests to add or remove one or more permissions for a role.

**Process Flow:**
1. Admin invokes Assign Permissions API (role id, list of permission identifiers or resource:action pairs).
2. System validates role exists and permissions are valid (optionally against Resource Registry).
3. System adds or replaces permission assignments for the role.
4. System records audit event.
5. System invalidates or updates any cached policy data for that role.
6. System returns updated role permissions.

**Output:** Updated list of permissions for the role; 200 on success.

**Constraints:** Permission format resource:action or wildcard; audit trail; cache invalidation on change (Epic 8).

---

#### FR-005: Assign Role to User

**Traceability:** Epic 1, E1-US3

**Description:** The system shall allow a System Admin to assign one or more Roles to a User so that the user gains the permissions of those roles (and inherited roles).

**Trigger:** System Admin requests to assign role(s) to a user.

**Process Flow:**
1. Admin invokes Assign Role to User API (user id, role id(s), optional scope).
2. System validates user and role(s) exist and belong to the same tenant.
3. If scope is supported, system validates scope (e.g. org, project).
4. System creates user-role assignment(s).
5. System records audit event.
6. System invalidates cache for affected subject.
7. System returns confirmation.

**Output:** Assignment(s) created; 201 or 200.

**Constraints:** User and role must be in same tenant; optional scope; audit trail; cache invalidation.

---

#### FR-006: Assign Role to Group

**Traceability:** Epic 1, E1-US3

**Description:** The system shall allow a System Admin to assign one or more Roles to a Group so that every user in that group effectively has those roles.

**Trigger:** System Admin requests to assign role(s) to a group.

**Process Flow:**
1. Admin invokes Assign Role to Group API (group id, role id(s), optional scope).
2. System validates group and role(s) exist and belong to the same tenant.
3. System creates group-role assignment(s).
4. System records audit event.
5. System invalidates cache for the group and any user that is a member.
6. System returns confirmation.

**Output:** Assignment(s) created; 201 or 200.

**Constraints:** Group and role same tenant; role applies to all current and future members; audit trail; cache invalidation.

---

#### FR-007: Role Inheritance

**Traceability:** Epic 1, E1-US4

**Description:** The system shall support role hierarchy so that a child role inherits all permissions of its parent(s). When checking permission, the system shall consider both the subject's direct roles and their parent roles.

**Trigger:** Admin defines parent role for a role; or system evaluates permission for a subject.

**Process Flow:**
1. **Definition:** Admin invokes Set Role Parent API (child role id, parent role id). System validates no circular reference and stores hierarchy.
2. **Evaluation:** When resolving permissions for a subject, system collects all roles (direct user roles + group roles + inherited). For each role, system includes permissions from that role and all ancestors. Permission check uses this expanded set.

**Output:** Hierarchy stored; permission evaluation includes inherited permissions.

**Constraints:** No circular inheritance; inheritance evaluated at permission check time; cache must account for hierarchy.

---

### Epic 2: Groups & Delegation

---

#### FR-008: Create and Manage Groups

**Traceability:** Epic 2, E2-US1

**Description:** The system shall allow a System Admin to create Groups (name, description) and update or delete them, so that users can be managed by department/team.

**Trigger:** System Admin creates or updates a group.

**Process Flow:**
1. Admin invokes Create Group API (name, description) with tenant context.
2. System validates name uniqueness per tenant (if required).
3. System creates the group.
4. System records audit event.
5. For update/delete, system performs validation and applies change; audit and cache invalidation as needed.

**Output:** Group resource (id, name, description); 201/200/204.

**Constraints:** Tenant-scoped; audit trail.

---

#### FR-009: Add or Remove Users from Group

**Traceability:** Epic 2, E2-US1

**Description:** The system shall allow a System Admin to add users to a Group and remove users from a Group. Roles assigned to the group apply to all members.

**Trigger:** System Admin requests to add or remove user(s) to/from a group.

**Process Flow:**
1. Admin invokes Add Users to Group or Remove Users from Group API (group id, user id(s)).
2. System validates group and users exist and belong to the same tenant.
3. System updates group membership.
4. System records audit event.
5. System invalidates cache for affected users (their effective permissions may have changed).
6. System returns confirmation.

**Output:** Updated membership; 200 or 204.

**Constraints:** Same-tenant; audit; cache invalidation for affected users.

---

#### FR-010: Delegation (Assign/Revoke Role by Delegated Admin)

**Traceability:** Epic 2, E2-US2

**Description:** The system shall support Delegation: User A can be authorized to assign or revoke a specific Role R for users within a defined scope S (e.g. a group or org), enabling delegated administration.

**Trigger:** System Admin creates a delegation rule; later, delegated user A assigns/revokes role R for a user in scope S.

**Process Flow:**
1. **Create delegation:** Admin defines delegation rule (delegator user A, role R, scope S, allow assign/revoke). System stores rule and records audit.
2. **Delegated assign/revoke:** User A invokes Assign/Revoke Role API for a target user. System checks that A has a delegation rule covering role R and that target user is in scope S. If authorized, system performs assignment/revoke, records audit (including delegator identity), invalidates cache.

**Output:** Delegation rule created; assignment/revoke result when delegated action is performed.

**Constraints:** Scope S must be enforceable (e.g. group id, org id); audit must record delegator; same tenant.

---

### Epic 3: Policy Decision & Enforcement

---

#### FR-011: Check Permission API

**Traceability:** Epic 3, E3-US1

**Description:** The system shall expose a Check Permission API that accepts subject, resource, and action and returns Allow or Deny, so that microservices can enforce authorization. Batch check shall be supported.

**Trigger:** Application (PEP) calls RBAC service to check if a subject is allowed to perform an action on a resource.

**Process Flow:**
1. Client sends `POST /check` (or equivalent) with body e.g. `{ subject, resource, action }` and tenant context.
2. System resolves subject to user (and optionally groups), resolves all roles (including group roles and inherited).
3. System collects all permissions from those roles and evaluates whether (resource, action) is permitted (including wildcard matching).
4. System logs the decision (Allow/Deny) for audit.
5. System returns `{ allowed: boolean }` (and optionally batch results if batch check).
6. Caching may serve result without full evaluation when valid cache entry exists.

**Output:** `{ allowed: true|false }` per check; for batch, array of results. Response time p99 < 50ms (NFR).

**Constraints:** Subject, resource, action required; tenant required; audit 100% of checks; latency SLA.

---

#### FR-012: SDK for Permission Check

**Traceability:** Epic 3, E3-US2

**Description:** The system shall provide SDKs (Node, Go, Java) with a simple method (e.g. `check(userId, resource, action)`) that calls the Check Permission API, with handling for cache, retry, and errors.

**Trigger:** Application developer integrates SDK and calls check from application code.

**Process Flow:**
1. Application calls SDK method with userId (or subject id), resource, action; SDK may accept tenant from context or parameter.
2. SDK may check local/sidecar cache first; on miss, calls RBAC Check Permission API.
3. SDK implements retry and timeout per contract; returns allowed/deny or throws on error.
4. Application uses result to allow or deny the operation.

**Output:** Boolean (or result object) indicating allowed/denied; integration achievable in &lt; 1 day per service (success metric).

**Constraints:** SDK must align with API contract; cache TTL and invalidation consistent with FR-023.

---

#### FR-013: PDP Standalone (Control and Data Plane)

**Traceability:** Epic 3, E3-US3

**Description:** The system shall implement a Policy Decision Point (PDP) as a standalone component, separate from policy administration, so that decision logic can scale independently and policy can be updated in real time without redeploying consuming services.

**Trigger:** Policy data is updated (roles, permissions, assignments); or a check request is received.

**Process Flow:**
1. **Control plane:** Admin/API updates roles, permissions, assignments; changes are persisted and propagated to PDP data plane (e.g. via cache refresh or event).
2. **Data plane:** PDP instances serve check requests from a local policy snapshot/cache; they do not perform CRUD on roles/permissions.
3. Policy updates are reflected in PDP in real time (or within configured propagation delay) so that new permissions take effect without service redeploy.

**Output:** Decoupled PDP that serves Allow/Deny from current policy; scalable and independently deployable.

**Constraints:** Stateless PDP instances; policy propagation mechanism; no direct DB write from PDP data plane for admin data.

---

### Epic 4: Multi-Tenancy & Scope

---

#### FR-014: Multi-Tenancy (Tenant Isolation)

**Traceability:** Epic 4, E4-US1

**Description:** The system shall support multi-tenancy so that all entities (users, roles, permissions, groups) are scoped to a tenant. Each tenant has its own RBAC model; API calls shall include tenant context (e.g. header `X-Tenant-ID`).

**Trigger:** Any API call that reads or writes RBAC entities.

**Process Flow:**
1. Every request carries tenant identifier (header, token claim, or routing).
2. All CRUD and check operations filter and store data by tenant id.
3. No data from one tenant is returned or applied in the context of another tenant.
4. Tenant id is validated (e.g. exists, caller allowed) where applicable.

**Output:** Full isolation of RBAC data per tenant; consistent tenant context in audit.

**Constraints:** Tenant id mandatory for all operations; no cross-tenant data leakage; compliance (GDPR, etc.) per tenant.

---

#### FR-015: Scope for Roles (Org/Project)

**Traceability:** Epic 4, E4-US2

**Description:** The system shall support Scope (e.g. org, project) so that a role assignment can be limited to a scope. When checking permission, the system shall consider scope (e.g. user has Admin role in org A but not in org B).

**Trigger:** Admin assigns role with scope; or permission check is performed with resource context including scope.

**Process Flow:**
1. **Assignment:** When assigning role to user/group, admin may specify scope (e.g. org_id, project_id). System stores assignment with scope.
2. **Check:** Client may pass scope context (e.g. org_id) with check request. System resolves roles for subject and filters assignments by scope; only roles matching the requested scope (or global) are considered. Permission is allowed only if (resource, action) is granted in that scope.

**Output:** Scope-aware assignments and scope-aware Allow/Deny decisions.

**Constraints:** Scope model (hierarchy, global vs scoped) defined; backward compatibility for requests without scope (e.g. treat as global).

---

### Epic 5: Audit & Compliance

---

#### FR-016: Audit Log for Permission Checks

**Traceability:** Epic 5, E5-US1

**Description:** The system shall record an audit log entry for every permission check (Allow or Deny), including user (subject), resource, action, timestamp, result, and tenant, to support compliance.

**Trigger:** Every invocation of Check Permission (API or via PDP).

**Process Flow:**
1. Before or after returning the decision, system writes an immutable audit record: subject, resource, action, timestamp, decision (Allow/Deny), tenant_id, optional request_id.
2. Log is stored in audit store with retention per policy (e.g. minimum 1 year configurable).
3. Log is available for query via Audit Query API.

**Output:** 100% of permission checks logged; queryable by user, resource, date range.

**Constraints:** No suppression of audit for performance; retain at least 1 year (configurable); GDPR/PCI-DSS/SOC2 alignment.

---

#### FR-017: Audit Log for RBAC Admin Actions

**Traceability:** Epic 5, E5-US2

**Description:** The system shall record an audit log entry for every RBAC administrative change (create/update/delete role, assign/revoke permission, assign/revoke role to user/group, group membership changes, delegation rules, etc.), including who made the change and when.

**Trigger:** Any successful (or attempted) admin API call that changes RBAC state.

**Process Flow:**
1. After each admin operation, system writes audit record: actor (admin/delegated user), action type, target entity (e.g. role id, user id), before/after or change description, timestamp, tenant.
2. Records are immutable and queryable.
3. Export and reporting can include these entries.

**Output:** Full traceability of who changed what and when; support for compliance and troubleshooting.

**Constraints:** All CRUD and assignment changes covered; immutable log; retention as per FR-016.

---

#### FR-018: Export Audit Log

**Traceability:** Epic 5, E5-US3

**Description:** The system shall provide an API to export audit log (CSV or JSON) for a time range and optional filters (user, resource, action, tenant), with pagination support, for internal reporting and compliance.

**Trigger:** Auditor or admin requests export (e.g. by date range, user, resource).

**Process Flow:**
1. Client calls Export Audit API with filters (date_from, date_to, user_id, resource, action, tenant) and format (CSV/JSON), optionally pagination (limit, offset or cursor).
2. System authorizes caller (e.g. Auditor role).
3. System queries audit store and returns data in requested format, respecting pagination.
4. Large exports may be async (job id + download link) if defined.

**Output:** Audit records in CSV or JSON; paginated; filtered by criteria.

**Constraints:** Authorization required; retention and data size may limit export range; compliance-safe format.

---

### Epic 6: Resource Registry & Policy as Code

---

#### FR-019: Register Resources and Actions

**Traceability:** Epic 6, E6-US1

**Description:** The system shall allow a Service Owner to register the Resources and Actions of their service (e.g. `order`, `payment` with actions `read`, `create`, `delete`), so that RBAC has a taxonomy and can validate that permissions reference registered resources.

**Trigger:** Service Owner registers or updates resource types and actions for their service.

**Process Flow:**
1. Service Owner invokes Register Resource API (e.g. resource type, list of actions) with tenant/service context.
2. System stores resource registry entry; may support versioning.
3. When permissions are assigned (FR-004), system may validate that resource:action exists in registry (optional strict mode).
4. Check Permission continues to resolve against assigned permissions; registry is for validation and discovery.

**Output:** Resource registry updated; optional validation of permission strings against registry.

**Constraints:** Idempotent registration; per-tenant or global registry depending on design.

---

#### FR-020: Policy as Code (Import from YAML/JSON)

**Traceability:** Epic 6, E6-US2

**Description:** The system shall support Policy as Code: importing policy definitions (roles, role-permissions, role assignments) from YAML or JSON files, with version control and CI/CD-friendly workflow.

**Trigger:** Admin or CI pipeline invokes import API with policy file or URL.

**Process Flow:**
1. Client sends policy document (YAML/JSON) defining roles, permissions per role, and optionally assignments.
2. System validates schema and references (tenant, users, groups).
3. System applies policy (create/update roles, permissions, assignments); may support dry-run or diff.
4. System records audit (e.g. "policy import", file or version id).
5. Cache invalidation and PDP refresh as for other admin changes.

**Output:** Policy state updated to match file; audit trail of import; optional diff/preview.

**Constraints:** Atomic or phased apply; rollback strategy; no conflict with concurrent admin UI changes (or define merge strategy).

---

### Epic 7: Built-in Roles & Default Policies

---

#### FR-021: Built-in Roles (Viewer, Editor, Admin)

**Traceability:** Epic 7, E7-US1

**Description:** The system shall provide built-in roles (e.g. Viewer, Editor, Admin) per resource type or globally per tenant, to speed up onboarding. These can be overridden or customized per tenant.

**Trigger:** Tenant onboarding or admin selects built-in role set; or admin customizes built-in role permissions.

**Process Flow:**
1. On tenant creation (or first use), system may create default roles (Viewer, Editor, Admin) with standard permission sets per registered resource types.
2. Admin can view and optionally edit built-in role permissions (without deleting the role).
3. Assigning built-in roles to users/groups follows FR-005, FR-006.
4. Custom roles remain available alongside built-in ones.

**Output:** Pre-defined roles available per tenant; consistent starting point; customizable within tenant.

**Constraints:** Built-in roles identifiable (e.g. system flag); prevent accidental deletion of built-in roles or define behavior.

---

#### FR-022: Default Policy for New Tenant

**Traceability:** Epic 7, E7-US2

**Description:** When a new tenant is created, the system shall apply a default policy (e.g. deny-all or a configurable template) so that the tenant is secure by default until explicitly granted permissions.

**Trigger:** New tenant creation.

**Process Flow:**
1. Tenant creation API or process creates tenant record.
2. System applies default policy: e.g. deny all permissions, or apply template (e.g. built-in roles with no assignments).
3. No role assignments exist until admin assigns roles; permission checks for that tenant return Deny until then.
4. Optional: configurable template per tenant type (e.g. strict vs. template with one admin role).

**Output:** New tenant in deny-all or template state; no implicit permissions.

**Constraints:** No default allow for sensitive resources; document template options.

---

### Epic 8: Performance & Scalability

---

#### FR-023: Caching for PDP

**Traceability:** Epic 8, E8-US1

**Description:** The system shall provide a cache layer (edge or sidecar) for permission decisions and/or policy data, with configurable TTL. Cache shall be invalidated when policy or assignments change so that results stay correct.

**Trigger:** Permission check (cache hit vs miss); or policy/assignment change (invalidation).

**Process Flow:**
1. On permission check, PDP or SDK checks cache (e.g. key: tenant+subject+resource+action). On hit, return cached Allow/Deny within latency budget.
2. On miss, full evaluation is performed; result is cached with TTL.
3. When roles, permissions, or assignments are modified (FR-001–FR-010, FR-020), system triggers invalidation for affected keys (e.g. by subject, role, or tenant).
4. TTL and invalidation strategy are configurable.

**Output:** Reduced latency for repeated checks; p99 < 50ms; correct results after invalidation.

**Constraints:** Consistency: invalidation on every relevant change; TTL as fallback; no stale allow for sensitive actions beyond acceptable window.

---

#### FR-024: Horizontal Scaling of RBAC Service

**Traceability:** Epic 8, E8-US2

**Description:** The RBAC service (especially PDP) shall scale horizontally: additional instances can be added behind a load balancer to handle increased traffic, with no per-instance state required for decision logic.

**Trigger:** Load increase or deployment of new instances.

**Process Flow:**
1. PDP and API components are stateless; any instance can serve any request.
2. Load balancer distributes requests across instances.
3. Shared data store and cache (or cache cluster) provide consistent policy and cache state.
4. Scaling is achieved by adding instances and routing traffic; p99 latency and availability targets maintained.

**Output:** Service meets 99.9% availability and p99 < 50ms under target load; linear scaling with instances.

**Constraints:** Stateless design; shared persistence and cache; no affinity required for permission checks.

---

## 4. Cross-Cutting Notes

- **Authentication:** All APIs assume an authenticated caller; identity (and optionally tenant) come from IdP (OAuth2/OIDC). RBAC does not authenticate users but uses subject id from token.
- **Authorization of admin APIs:** System Admin and delegated actions (FR-010) must be protected by RBAC or equivalent so that only authorized users can perform admin operations.
- **Idempotency:** Create/Update of roles, groups, and registry may support idempotency keys for safe retries.

---

## 5. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | Business BA | Initial FRS from PRD and Epic brief |
