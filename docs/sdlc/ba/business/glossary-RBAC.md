# Glossary — RBAC for Microservices

**Version:** 1.0  
**Date:** 2025-03-15  
**Author:** Business BA  
**Source:** PRD §9, expanded for FRS and process flows

---

## A–C

| Term | Definition |
|------|------------|
| **Action** | An operation that can be performed on a resource (e.g. `read`, `create`, `delete`, `approve`). Combined with a resource to form a permission (e.g. `order:read`). |
| **Audit log** | Immutable record of events for compliance and traceability. In RBAC: (1) every permission check (Allow/Deny) with subject, resource, action, timestamp; (2) every administrative change (CRUD roles, assign/revoke, delegation). Retained per policy (e.g. minimum 1 year). |
| **Auditor** | User persona who reviews access and compliance: queries and exports audit logs, produces reports. |
| **Authorization** | Decision whether a subject is allowed to perform an action on a resource. RBAC provides authorization (Allow/Deny); authentication (who you are) is provided by IdP. |
| **Batch check** | Single API call that evaluates multiple permission checks (e.g. multiple subject–resource–action triples) and returns an array of Allow/Deny results. |
| **Built-in role** | Pre-defined role (e.g. Viewer, Editor, Admin) provided per tenant or per resource type to speed onboarding; may be customized but not deleted. |
| **Control plane** | Part of the RBAC system that manages policy: CRUD roles, permissions, assignments, delegation. Distinct from the data plane (PDP) that serves check requests. |
| **Delegation** | Mechanism by which a user (delegator) is allowed to assign or revoke a specific role for users within a defined scope (e.g. group, org), without being a full system admin. |

---

## D–G

| Term | Definition |
|------|------------|
| **Data plane** | Part of the RBAC system that serves permission check requests (PDP). Stateless, scalable; reads policy from cache/store, does not perform admin CRUD. |
| **Default policy** | Policy applied when a new tenant is created (e.g. deny-all or a template). Ensures secure-by-default until roles are assigned. |
| **Enforcement** | Application of the authorization decision at the point of use (e.g. blocking or allowing an API call). Done by the PEP in the consuming service. |
| **Epic** | Large feature or theme in the PRD; groups related user stories (e.g. Epic 1: Core RBAC Model). |
| **FRS** | Functional Requirements Specification; detailed requirements derived from PRD Epics and User Stories. |
| **Group** | Collection of users within a tenant (e.g. department, team). Roles can be assigned to a group; all members inherit those roles. |
| **Group-role assignment** | Link between a group and a role; every member of the group receives the permissions of that role (within tenant/scope). |

---

## I–P

| Term | Definition |
|------|------------|
| **IdP** | Identity Provider. Supplies user identity (e.g. OAuth2/OIDC); RBAC consumes subject id (e.g. `sub`) from token, does not authenticate. |
| **Inheritance (role)** | Hierarchy where a child role inherits all permissions of its parent(s). When resolving permissions for a subject, the system considers the subject’s roles and their ancestors. |
| **Multi-tenancy** | Isolation of data by tenant. Each tenant has its own roles, permissions, users, groups; no cross-tenant data exposure. API requests carry tenant id (e.g. `X-Tenant-ID`). |
| **PDP** | Policy Decision Point. Component that evaluates policy and returns Allow or Deny for a given (subject, resource, action). In this system, PDP is standalone and scalable. |
| **PEP** | Policy Enforcement Point. Point in the application (e.g. microservice) where the authorization decision is enforced; the PEP calls the PDP (or SDK) to get the decision. |
| **Permission** | A pair (Resource, Action), e.g. `order:read`, `user:delete`. Permissions are assigned to roles; subjects get permissions through their roles (and inheritance). Wildcard (e.g. `*`) may be supported. |
| **Policy** | Set of roles, permissions, and assignments that define who can do what. “Policy as code” means policy defined in YAML/JSON and imported (e.g. from Git). |
| **Policy as Code** | Defining RBAC policy in files (YAML/JSON), versioned in Git, and importing/deploying via API or CI/CD. |
| **PRD** | Product Requirements Document. Source of Epics and User Stories for this project. |

---

## R–S

| Term | Definition |
|------|------------|
| **Resource** | An entity or asset that can be acted upon (e.g. `order`, `user`, `payment`). Part of the permission pair (resource:action). May be registered in the Resource Registry. |
| **Resource Registry** | Catalog of resource types and their allowed actions, registered by Service Owners. Used to validate permission strings and provide taxonomy. |
| **Role** | Named set of permissions representing a job function (e.g. Order Manager, Viewer). Roles are assigned to users or groups; users get the union of permissions from all their roles (and inherited roles). |
| **Role hierarchy** | Parent–child relationship between roles; child inherits parent’s permissions. No circular inheritance. |
| **Scope** | Optional context that limits where a role applies (e.g. org_id, project_id). User may have role “Admin” in org A but not in org B. Permission check considers scope when provided. |
| **SDK** | Software development kit. In RBAC: client libraries (Node, Go, Java) that expose a simple method (e.g. `check(userId, resource, action)`) and handle API calls, cache, retry. |
| **Service Owner** | Persona who owns a microservice; registers that service’s resources and actions in the Resource Registry. |
| **Subject** | The entity whose permission is being checked: typically a user (user_id) or a service account. |
| **System Admin** | Persona who manages the RBAC system: CRUD roles, assign permissions, assign roles to users/groups, manage groups and delegation, view audit. |

---

## T–Z

| Term | Definition |
|------|------------|
| **Tenant** | Unit of isolation; each tenant has its own RBAC model (users, roles, permissions, groups). Used in SaaS and multi-tenant architectures. |
| **User** | Identity from IdP; can be assigned roles directly or via group membership. Identified by user_id (e.g. from token `sub`). |
| **User story (US)** | Short requirement in form “As [persona], I want [capability], so that [benefit].” From PRD Epics; traced to FRS. |
| **Wildcard** | Placeholder in permission (e.g. `*`) meaning “any” resource or action (e.g. `order:*` = any action on order). Exact semantics defined in FRS/implementation. |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-03-15 | Business BA | Initial glossary (PRD terms + FRS/process flow terms) |
