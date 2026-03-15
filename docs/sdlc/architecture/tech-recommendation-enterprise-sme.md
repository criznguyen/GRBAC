# Đề xuất Công nghệ: Enterprise + SME

**Version:** 1.0  
**Date:** 2025-03-15  
**Mục tiêu:** Chọn stack phục vụ cả **enterprise-grade** và **SME** — scale-down (rẻ, đơn giản) và scale-up (HA, compliance) với cùng codebase.

---

## 1. Nguyên tắc lựa chọn

| Tiêu chí | Enterprise | SME | Chiến lược |
|----------|------------|-----|------------|
| **Cost** | Cao, ưu tiên HA | Thấp, ưu tiên tối giản | Cùng stack, khác deployment tier |
| **Scale** | Nhiều tenant, high throughput | Ít tenant, vừa phải | Stateless, add instance khi cần |
| **Ops** | Team DevOps, K8s | 1–2 người, Docker Compose | Hỗ trợ cả Compose và K8s |
| **Compliance** | Audit, multi-region | Audit cơ bản | Audit bắt buộc cho cả hai |
| **Time-to-market** | Chấp nhận dài hơn | Cần nhanh | Dev nhanh, deploy đơn giản |

**Nguyên tắc vàng:** Chọn công nghệ **scale-down tốt** (chạy được 1 instance) nhưng **scale-up dễ** (thêm replica, cluster). Tránh công nghệ chỉ phù hợp khi scale lớn.

---

## 2. Đề xuất chính

### 2.1 Runtime / Ngôn ngữ: **Go**

| Lý do | Enterprise | SME |
|-------|------------|-----|
| **Single binary** | Deploy dễ, ít dependency | Không cần runtime phức tạp |
| **Performance** | p99 < 50ms dễ đạt, ít CPU/RAM | Chạy ổn trên VPS nhỏ |
| **Concurrency** | Goroutines cho async audit, batch | Same |
| **Ecosystem** | ChiChi, Echo, gRPC chuẩn | Same |
| **Hiring** | Go phổ biến enterprise | Dễ tuyển, học nhanh |

**Thay thế xem xét:**
- **Node.js/TypeScript:** Dev nhanh hơn, nhiều dev hơn; nhưng memory cao hơn, p99 khó kiểm soát hơn khi load lớn. Phù hợp nếu team chủ yếu JS/TS.
- **Java/Spring Boot:** Tốt cho enterprise, nhưng footprint lớn, startup chậm — kém phù hợp SME single-node.

**Kết luận:** **Go** là lựa chọn cân bằng tốt nhất cho cả hai phân khúc.

---

### 2.2 Database: **PostgreSQL**

| Lý do | Enterprise | SME |
|-------|------------|-----|
| **ACID, reliability** | Chuẩn compliance | Chuẩn production |
| **Scale down** | 1 instance đủ cho SME | PostgreSQL chạy tốt trên 1–2 CPU |
| **Scale up** | Replication, read replicas | Cùng schema, thêm replica khi cần |
| **Managed options** | AWS RDS, Cloud SQL, Neon | Supabase, Neon free tier, Railway |
| **Partitioning** | Audit table partition theo tháng | Same schema |

**SME tip:** Dùng **Neon**, **Supabase**, hoặc **Railway** — free/cheap tier, managed backup, ít ops.

**Enterprise tip:** RDS / Cloud SQL với Multi-AZ, read replicas, point-in-time recovery.

---

### 2.3 Cache: **Redis** (có fallback in-memory cho SME)

| Tier | Deployment | Ghi chú |
|------|------------|---------|
| **SME** | Redis single instance (Docker) hoặc **in-memory fallback** | Nếu không có Redis: dùng LRU in-process, TTL ngắn (30s). Đủ cho vài trăm user. |
| **Enterprise** | Redis Sentinel / Cluster | HA, shared across PDP instances |

**Quyết định:** Redis là dependency **optional** ở MVP. Code hỗ trợ:
1. **Redis** (preferred) — shared cache, invalidation đúng
2. **In-memory** — fallback khi `REDIS_URL` empty; mất cache khi restart, không shared giữa instances. Chỉ dùng cho SME single-node.

---

### 2.4 API Framework: **Chi** (Go) hoặc **Echo**

| Framework | Ưu điểm |
|-----------|---------|
| **Chi** | Nhẹ, chuẩn net/http, middleware đơn giản, phổ biến |
| **Echo** | Tương tự, validation built-in, performance tốt |

Cả hai đều phù hợp. **Chi** đơn giản hơn, **Echo** có sẵn nhiều helper.

---

### 2.5 ORM / DB Layer: **sqlx** hoặc **sqlc**

| Option | Ưu điểm | Nhược điểm |
|--------|---------|------------|
| **sqlx** | Lightweight, raw SQL + struct mapping | Phải viết SQL tay |
| **sqlc** | Generate type-safe code từ SQL | Thêm bước generate |
| **GORM** | ORM đầy đủ | Nặng, dễ N+1, migration phức tạp |

**Đề xuất:** **sqlc** — type-safe, performance tốt, migrations tách (Atlas hoặc golang-migrate).

---

### 2.6 Deployment

| Tier | Option | Ghi chú |
|------|--------|---------|
| **SME** | **Docker Compose** | 1 container API+PDP, 1 PostgreSQL, 0–1 Redis. Chạy trên VPS $10–20/tháng. |
| **SME+** | **Railway / Render / Fly.io** | Managed PostgreSQL + Redis, deploy từ Git. |
| **Enterprise** | **Kubernetes** | Admin API và PDP tách service; HPA; Ingress; managed DB (RDS/Cloud SQL). |

**Quan trọng:** Cùng image Docker, khác `docker-compose.yml` vs K8s manifests. Không cần hai codebase.

---

### 2.7 IdP Integration

| IdP | SME | Enterprise |
|-----|-----|------------|
| **Auth0** | Free tier 7k MAU | Team plan, enterprise SSO |
| **Keycloak** | Self-hosted free | Self-hosted, cluster |
| **Supabase Auth** | Free tier | — |
| **Clerk** | Free tier | — |
| **Custom OAuth2** | Bất kỳ IdP chuẩn | Same |

RBAC **không** làm authentication — chỉ nhận JWT, validate, extract `sub` + `tenant`. Tích hợp IdP qua thư viện chuẩn (ví dụ `golang.org/x/oauth2`, `github.com/golang-jwt/jwt`).

---

## 3. So sánh theo tier

| Thành phần | SME (Tier 1) | SME+ (Tier 2) | Enterprise (Tier 3) |
|------------|--------------|---------------|---------------------|
| **Runtime** | Go | Go | Go |
| **API** | Chi/Echo | Chi/Echo | Chi/Echo |
| **DB** | PostgreSQL (Neon/Supabase free) | PostgreSQL (Railway/Render) | PostgreSQL (RDS Multi-AZ) |
| **Cache** | In-memory fallback | Redis single | Redis Sentinel/Cluster |
| **Deploy** | Docker Compose | Railway/Render | Kubernetes |
| **Cost/tháng** | ~$0–20 | ~$25–50 | $200+ |
| **Users** | < 500 | 500–5k | 5k+ |
| **Tenants** | 1–10 | 10–50 | 50+ |

---

## 4. Migration path

1. **Phase 1 (MVP):** Tier 1 — Docker Compose, in-memory cache, 1 PostgreSQL.
2. **Phase 2:** Thêm Redis khi cần scale hoặc multi-instance.
3. **Phase 3:** Chuyển sang Tier 2 (managed) khi ops trở nên nặng.
4. **Phase 4:** Tier 3 (K8s, HA) khi có yêu cầu enterprise.

Code không đổi — chỉ thay đổi config và deployment.

---

## 5. Cập nhật tech-stack.md

Đề xuất cập nhật `tech-stack.md`:

- **Language:** Go 1.22+
- **Framework:** Chi hoặc Echo
- **DB layer:** sqlc
- **Migrations:** golang-migrate hoặc Atlas
- **Cache:** Redis (primary); in-memory fallback cho SME single-node
- **Policy Store / Audit Store:** PostgreSQL (có thể 1 DB cho SME, 2 DB cho Enterprise)

---

## 6. Kết luận

| Thành phần | Đề xuất |
|------------|---------|
| **Runtime** | **Go** |
| **Framework** | Chi hoặc Echo |
| **Database** | PostgreSQL |
| **Cache** | Redis (+ in-memory fallback) |
| **Deploy SME** | Docker Compose |
| **Deploy Enterprise** | Kubernetes |

Stack này cho phép:
- **SME:** Deploy nhanh, chi phí thấp, ít phụ thuộc, vẫn đủ cho production.
- **Enterprise:** Scale, HA, compliance, cùng codebase và mô hình dữ liệu.
