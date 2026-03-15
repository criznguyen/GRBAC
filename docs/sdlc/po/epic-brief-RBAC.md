# Epic Brief: RBAC for Microservices

## Problem

Các hệ thống microservices thiếu cơ chế authorization thống nhất: logic phân quyền trùng lặp, không có single source of truth, khó audit và compliance. Khách hàng cần RBAC service có thể áp dụng ngay vào production.

## Success Metrics

- Time to integrate: < 1 ngày/service
- Latency p99: < 50ms
- Availability: 99.9%
- Audit compliance: 100% (mọi access có trail)

## User Stories (Summary)

1. **Core RBAC:** Tạo/sửa role, gán permission cho role, gán role cho user/group, role inheritance
2. **Groups & Delegation:** Quản lý group, ủy quyền quản lý role assignment
3. **Policy Decision:** Check Permission API, SDK, PDP standalone
4. **Multi-Tenancy:** Isolation theo tenant, scope (org/project)
5. **Audit:** Log mọi check + admin actions, export
6. **Resource Registry:** Đăng ký resource/action, policy as code
7. **Built-in Roles:** Viewer/Editor/Admin, default policy
8. **Performance:** Caching, horizontal scaling

## Acceptance Criteria (High-level)

- [ ] RBAC service expose API check permission (Allow/Deny)
- [ ] Hỗ trợ CRUD Roles, Permissions, Users, Groups
- [ ] Role inheritance, multi-tenancy
- [ ] Audit log đầy đủ, export được
- [ ] SDK cho Node, Go, Java
- [ ] p99 < 50ms, 99.9% availability

## Priority

**Must have:** Epic 1, 3, 4, 5, 8  
**Should have:** Epic 2, 6, 7  
**Could have:** Epic 6-US2 (Policy as Code)

## Dependencies & Risks

- **Dependencies:** IdP (OAuth2/OIDC) cho user identity
- **Risks:** Nếu IdP chậm, có thể ảnh hưởng latency — mitigate bằng cache user context

## Handoff

**→ Business BA** — Chi tiết hóa thành FRS, process flows, use cases tại `docs/sdlc/ba/business/`
