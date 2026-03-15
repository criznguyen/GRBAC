# Product Requirements Document (PRD)
# Hệ thống RBAC cho Microservices

**Version:** 1.0  
**Date:** 2025-03-15  
**Author:** Senior PO (10+ năm kinh nghiệm)  
**Status:** Draft → Ready for Business BA

---

## 1. Executive Summary

Khách hàng cần phát triển **hệ thống RBAC (Role-Based Access Control)** phục vụ kiến trúc microservices, với đầy đủ tính năng enterprise-grade để có thể áp dụng ngay vào các hệ thống sản xuất. Hệ thống phải hỗ trợ nhiều service, nhiều tenant, và có thể mở rộng khi nhu cầu tăng.

---

## 2. Business Context & Problem Statement

### 2.1 Vấn đề
- Các hệ thống microservices hiện tại thiếu cơ chế authorization thống nhất, dẫn đến:
  - Logic phân quyền trùng lặp, khó bảo trì
  - Không có single source of truth cho roles và permissions
  - Khó audit, khó compliance (GDPR, PCI-DSS, SOC2)
  - Onboarding/offboarding người dùng phức tạp

### 2.2 Mục tiêu
- Cung cấp **RBAC service** trung tâm, có thể tích hợp nhanh vào mọi microservice
- Hỗ trợ đầy đủ tính năng RBAC chuẩn industry
- Sẵn sàng áp dụng production, hỗ trợ scale và multi-tenancy

---

## 3. Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Time to integrate | < 1 ngày/service | Số ngày từ lúc đọc docs đến khi service gọi RBAC |
| Latency p99 | < 50ms | Authorization check end-to-end |
| Availability | 99.9% | Uptime của RBAC service |
| Audit compliance | 100% | Mọi access đều có audit trail |

---

## 4. Scope & Out of Scope

### In Scope
- RBAC core (Users, Roles, Permissions, Groups)
- Role inheritance, delegation
- Policy decision point (PDP) standalone
- Audit logging, reporting
- Multi-tenancy
- SDK/API cho microservices

### Out of Scope (Phase 1)
- Authentication (OAuth2/OIDC) — tích hợp với IdP có sẵn
- ABAC/ReBAC đầy đủ — có thể mở rộng sau
- UI admin — có thể làm phase 2

---

## 5. User Personas

| Persona | Mô tả | Nhu cầu chính |
|---------|-------|---------------|
| **System Admin** | Quản lý toàn hệ thống | CRUD roles, assign permissions, xem audit |
| **Service Owner** | Sở hữu microservice | Đăng ký resource, định nghĩa permissions cho service |
| **Application Developer** | Tích hợp RBAC vào service | Gọi API check permission, sử dụng SDK |
| **Auditor** | Kiểm tra compliance | Xem audit log, báo cáo truy cập |

---

## 6. Epics & User Stories

### Epic 1: Core RBAC Model (Must Have)

| ID | User Story | Acceptance Criteria | Priority |
|----|------------|---------------------|----------|
| E1-US1 | As System Admin, I want tạo/sửa/xóa **Roles** với tên và mô tả, để nhóm permissions theo chức năng | Role có name, description, status; CRUD đầy đủ; validation unique name per tenant | Must |
| E1-US2 | As System Admin, I want gán **Permissions** cho Roles, để kiểm soát ai được làm gì | Permission = resource + action (e.g. `order:read`, `user:delete`); hỗ trợ wildcard `*` | Must |
| E1-US3 | As System Admin, I want gán **Roles** cho Users (hoặc Groups), để quản lý quyền tập trung | User có thể có nhiều role; Group có nhiều user; role gán cho group áp dụng cho tất cả user trong group | Must |
| E1-US4 | As System Admin, I want **Role inheritance** (role con kế thừa permissions của role cha), để giảm cấu hình trùng lặp | Role hierarchy; khi check permission, kiểm tra cả role và role cha | Must |

### Epic 2: Groups & Delegation (Should Have)

| ID | User Story | Acceptance Criteria | Priority |
|----|------------|---------------------|----------|
| E2-US1 | As System Admin, I want tạo **Groups** và gán Users vào Groups, để quản lý quyền theo phòng ban/team | Group có name, description; add/remove users; role gán cho group áp dụng cho mọi user | Should |
| E2-US2 | As System Admin, I want **Delegation** — ủy quyền user A quản lý role assignment cho nhóm user B, để phân cấp quản trị | Delegation rule: A có thể assign/revoke role R cho users trong scope S | Should |

### Epic 3: Policy Decision & Enforcement (Must Have)

| ID | User Story | Acceptance Criteria | Priority |
|----|------------|---------------------|----------|
| E3-US1 | As Application Developer, I want gọi **Check Permission API** (user, resource, action) và nhận Allow/Deny, để enforce trong service | API: `POST /check` body `{subject, resource, action}` → `{allowed: bool}`; hỗ trợ batch check | Must |
| E3-US2 | As Application Developer, I want sử dụng **SDK** (Node, Go, Java) để check permission với 1 dòng code, để tích hợp nhanh | SDK có method `check(userId, resource, action)`; xử lý cache, retry | Must |
| E3-US3 | As System, RBAC service phải có **PDP standalone** (tách biệt enforcement), để scale độc lập và cập nhật policy không cần deploy service | PDP có control plane và data plane riêng; policy cập nhật real-time | Must |

### Epic 4: Multi-Tenancy & Scope (Must Have)

| ID | User Story | Acceptance Criteria | Priority |
|----|------------|---------------------|----------|
| E4-US1 | As System Admin, I want **Multi-tenancy** — mỗi tenant có roles/permissions riêng, để hỗ trợ SaaS | Mọi entity (user, role, permission) thuộc tenant; API có header `X-Tenant-ID` | Must |
| E4-US2 | As System Admin, I want **Scope** (e.g. org, project) để giới hạn phạm vi role, để fine-grained control | Role có scope; check permission xét scope (e.g. user có role Admin trong org A, không có trong org B) | Should |

### Epic 5: Audit & Compliance (Must Have)

| ID | User Story | Acceptance Criteria | Priority |
|----|------------|---------------------|----------|
| E5-US1 | As Auditor, I want **Audit log** ghi lại mọi check permission (Allow/Deny) với user, resource, action, timestamp, để phục vụ compliance | Mọi check đều log; có API query audit theo user, resource, date range | Must |
| E5-US2 | As Auditor, I want **Audit log** ghi lại mọi thay đổi RBAC (CRUD role, assign/revoke), để traceability | Admin actions (create role, assign permission, etc.) đều có audit trail | Must |
| E5-US3 | As Auditor, I want export audit log (CSV/JSON) theo khoảng thời gian, để báo cáo nội bộ | API export với filters; hỗ trợ pagination | Should |

### Epic 6: Resource Registry & Policy as Code (Should Have)

| ID | User Story | Acceptance Criteria | Priority |
|----|------------|---------------------|----------|
| E6-US1 | As Service Owner, I want đăng ký **Resources** và **Actions** của service mình (e.g. `order`, `payment`), để RBAC biết taxonomy | API đăng ký resource type + actions; RBAC validate permission thuộc resource đã đăng ký | Should |
| E6-US2 | As System Admin, I want **Policy as Code** — định nghĩa policy bằng YAML/JSON, version control trong Git, để quản lý policy như infrastructure | Import policy từ file; support CI/CD deploy policy | Could |

### Epic 7: Built-in Roles & Default Policies (Should Have)

| ID | User Story | Acceptance Criteria | Priority |
|----|------------|---------------------|----------|
| E7-US1 | As System Admin, I want **Built-in roles** (Viewer, Editor, Admin) cho mỗi resource type, để onboarding nhanh | Pre-defined roles per tenant; có thể override/customize | Should |
| E7-US2 | As System, khi tạo tenant mới, có **default policy** (deny all, hoặc template), để secure by default | Tenant mới bắt đầu với deny-all hoặc template tùy config | Should |

### Epic 8: Performance & Scalability (Must Have)

| ID | User Story | Acceptance Criteria | Priority |
|----|------------|---------------------|----------|
| E8-US1 | As Application Developer, I want RBAC support **caching** tại edge/sidecar để giảm latency, để không làm chậm request | PDP có cache layer; TTL configurable; invalidate on policy change | Must |
| E8-US2 | As System, RBAC phải **scale horizontally** — thêm instance khi load tăng, để đáp ứng traffic | Stateless design; support load balancer; p99 < 50ms | Must |

---

## 7. Non-Functional Requirements

| NFR | Requirement |
|-----|-------------|
| Security | Zero-trust; least privilege; support JWT/OAuth2 token validation |
| Latency | p99 < 50ms cho check permission |
| Availability | 99.9% SLA |
| Data | Audit log retain tối thiểu 1 năm (configurable) |
| Compliance | Hỗ trợ GDPR, PCI-DSS, SOC2 audit requirements |

---

## 8. Dependencies & Assumptions

### Dependencies
- Identity Provider (IdP) cung cấp user identity (OAuth2/OIDC) — RBAC nhận `user_id`/`sub` từ token
- Message queue (optional) cho event-driven cache invalidation

### Assumptions
- Microservices gọi RBAC qua HTTP/gRPC (sync) hoặc sidecar
- Tenant ID có sẵn từ context (header, token claim, hoặc routing)

---

## 9. Glossary

| Term | Definition |
|------|------------|
| **Role** | Tập hợp permissions; đại diện cho job function (e.g. Order Manager) |
| **Permission** | Cặp (Resource, Action) — e.g. `order:create`, `user:delete` |
| **Subject** | User hoặc Service account cần check permission |
| **PDP** | Policy Decision Point — thành phần quyết định Allow/Deny |
| **PEP** | Policy Enforcement Point — điểm trong ứng dụng gọi PDP |
| **Tenant** | Đơn vị isolation — mỗi tenant có RBAC model riêng |

---

## 10. Handoff to Business BA

**Next Phase:** Business BA  
**Input:** PRD này + Epic briefs  
**Output expected:** Functional Requirements Specification (FRS), process flows, use cases, glossary  
**Output location:** `docs/sdlc/ba/business/`

**Instructions for Business BA:**
1. Phân rã mỗi Epic thành Functional Requirements chi tiết
2. Vẽ process flows cho các use case chính: Assign Role, Check Permission, Audit Query
3. Bổ sung use case diagram (textual) và glossary chi tiết
4. Đảm bảo FRS traceable về Epic/US trong PRD
