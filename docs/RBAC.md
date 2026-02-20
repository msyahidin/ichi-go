# RBAC System

Role-based access control with multi-tenant support, Casbin v3, and compliance-grade audit logging.

---

## Architecture Overview

### Three-Layer Permission Model

```
┌────────────────────────────────────────────┐
│  Layer 1: Platform Permissions (Global)    │
│  e.g. platform:admin:access               │
│  Cross-tenant operations, system admin     │
└──────────────────┬─────────────────────────┘
                   │
┌──────────────────▼─────────────────────────┐
│  Layer 2: Application Permissions          │
│  e.g. users:edit in tenant "acme"          │
│  Standard RBAC via Casbin, tenant-scoped   │
└──────────────────┬─────────────────────────┘
                   │
┌──────────────────▼─────────────────────────┐
│  Layer 3: Resource Permissions (Future)    │
│  e.g. product:123:edit                     │
│  Fine-grained ABAC (feature-flagged)       │
└────────────────────────────────────────────┘
```

### Key Components

| Component | Location | Purpose |
|-----------|----------|---------|
| RBAC domain | `internal/applications/rbac/` | 5 services, 5 controllers, 36 endpoints |
| Casbin enforcer | `internal/infra/authz/enforcer/` | Permission evaluation with caching |
| Bun adapter | `internal/infra/authz/adapter/` | Persists Casbin policies to MySQL |
| Cache | `internal/infra/authz/cache/` | L1 memory + L2 Redis two-tier cache |
| Watcher | `internal/infra/authz/watcher/` | RabbitMQ-based cache invalidation events |
| Circuit breaker | `internal/infra/authz/circuit_breaker/` | DB failure resilience |
| Tenant middleware | `internal/middlewares/tenant_context_middleware.go` | Resolves tenant from request |
| Enforcement middleware | `internal/middlewares/rbac_enforcement_middleware.go` | Checks permissions on routes |
| Audit middleware | `internal/middlewares/rbac_audit_middleware.go` | Logs operations for compliance |

### Casbin Model

The model file at `config/rbac_model.conf` defines:

- **Request definition**: `r = sub, dom, obj, act` (subject, tenant domain, resource, action)
- **Policy definition**: `p = sub, dom, obj, act` (maps roles to permissions)
- **Role groupings**:
  - `g` — tenant-scoped roles (role within a specific tenant)
  - `g2` — platform-global roles (role that applies across all tenants, `dom = *`)
- **Wildcard support**: `*` matches any tenant, resource, or action

---

## Configuration

```yaml
# config.local.yaml
rbac:
  # Deployment mode:
  # - "multi_tenant": Strict isolation, tenant_id required in every request
  # - "single_tenant": Auto-inject default tenant, best for internal admin tools
  # - "hybrid": Optional tenant_id with fallback to default (RECOMMENDED)
  mode: "hybrid"

  # Fallback tenant used in single_tenant and hybrid modes
  default_tenant: "system"

  # Path to Casbin model configuration
  model_path: "config/rbac_model.conf"

  performance:
    # "full": Load all policies into memory (small deployments)
    # "filtered": Load only per-tenant policies (recommended for multi-tenant)
    # "adaptive": Auto-switch based on policy count
    loading_strategy: "filtered"
    max_concurrent: 50

  cache:
    enabled: true
    memory_ttl: "60s"    # L1 in-memory LRU TTL
    redis_ttl: "5m"      # L2 Redis TTL (with LZ4 compression)
    max_size: 10000      # Max entries in L1 cache
    compression: true

  audit:
    enabled: true
    log_decisions: false  # High volume; use only for debugging
    log_mutations: true   # Recommended for compliance
```

---

## System Roles (Pre-seeded)

Nine roles are created automatically when you run `make seed-run`:

| Role | Slug | Scope | Description |
|------|------|-------|-------------|
| Super Admin | `super-admin` | Platform + all tenants | Full system access |
| Platform Admin | `platform-admin` | Platform | Platform-level operations |
| Tenant Admin | `tenant-admin` | Tenant | Full access within a tenant |
| Tenant Editor | `tenant-editor` | Tenant | Edit permissions in tenant |
| Tenant Viewer | `tenant-viewer` | Tenant | Read-only within tenant |
| Developer | `developer` | Tenant | Development tools access |
| Support | `support` | Tenant | Support operations |
| Auditor | `auditor` | Tenant | Audit log access only |
| Guest | `guest` | Tenant | Minimal access |

**Migration reference**: `db/migrations/schema/20260117_001_create_rbac_casbin_tables.sql`

---

## Usage Examples

### 1. Protect Routes with Middleware

**Option A: Global enforcement** — auto-maps HTTP methods to actions (GET→view, POST→create, etc.)

```go
import "ichi-go/internal/middlewares"

// In your route setup
enforcementService := do.MustInvoke[*rbacServices.EnforcementService](injector)

// Apply to all routes under a group
adminGroup := e.Group("/admin")
adminGroup.Use(middlewares.RBACEnforcementMiddleware(
    enforcementService,
    middlewares.DefaultRBACConfig(),
))
```

**Option B: Explicit permission on a specific route**

```go
// Protect a single route with a named permission
adminGroup.GET("/dashboard",
    myHandler,
    middlewares.RequirePermission(enforcementService, "admin", "access"),
)
```

### 2. Check Permission in Service Code

```go
// Inject enforcement service
enforcementService := do.MustInvoke[*services.EnforcementService](injector)

// Single check
allowed, err := enforcementService.CheckPermission(
    ctx,
    userID,    // int64
    tenantID,  // string, e.g. "acme-corp"
    "products", // resource
    "edit",     // action
)
if !allowed {
    return echo.NewHTTPError(http.StatusForbidden, "Access denied")
}

// Batch check (more efficient than multiple single checks)
checks := []services.PermissionCheck{
    {Resource: "products", Action: "view"},
    {Resource: "orders", Action: "view"},
}
results, err := enforcementService.CheckBatch(ctx, userID, tenantID, checks)
// results["products:view"] = true
// results["orders:view"] = false
```

### 3. Assign a Role to a User

```go
userRoleService := do.MustInvoke[*services.UserRoleService](injector)

err := userRoleService.AssignRole(
    ctx,
    userID,             // int64 — the user to assign the role to
    "tenant-editor",    // role slug
    "acme-corp",        // tenant ID
    adminID,            // int64 — who is making the assignment
    "Promoted to editor", // reason (saved to audit log)
)
```

### 4. Read Tenant Context in a Handler

```go
func myHandler(c echo.Context) error {
    ctx := c.Request().Context()

    userID := requestctx.GetUserIDAsInt64(ctx)
    tenantID := requestctx.GetTenantID(ctx)  // set by TenantContextMiddleware

    // ... your handler logic
    return c.JSON(http.StatusOK, result)
}
```

---

## Multi-Tenant Concepts

A **tenant** is any string identifier used to isolate a customer or organization's data — for example `"acme-corp"`, `"tenant-123"`, or `"default"`.

Casbin uses the tenant as the **domain** parameter, so a user's role in `"acme-corp"` has no effect in `"globex-corp"`.

**Resolving the tenant from the request** is handled by `TenantContextMiddleware`. Configure the strategy in your route setup:

```go
import "ichi-go/internal/middlewares"

// "auto" tries header → subdomain → path → query in order
e.Use(middlewares.TenantContextMiddleware(middlewares.TenantConfig{
    Strategy:       "auto",
    DefaultTenant:  "default",
    RequireTenant:  false,
}))

// Header only (most common for APIs)
e.Use(middlewares.TenantContextMiddleware(middlewares.TenantConfig{
    Strategy:   "header",
    HeaderName: "X-Tenant-Id", // default
}))
```

**Choosing a deployment mode:**

| Mode | When to use |
|------|-------------|
| `multi_tenant` | SaaS product — each customer is fully isolated |
| `single_tenant` | Internal tool — one team, no isolation needed |
| `hybrid` | Mixed: admin tools + multi-tenant APIs in the same app |

---

## Performance

Two-tier caching is built-in. The default hit rates under normal load:

| Cache Layer | Hit Rate | Latency | Config Key |
|-------------|----------|---------|------------|
| L1 Memory (LRU) | 80–90% | < 1ms | `rbac.cache.memory_ttl` |
| L2 Redis (LZ4) | 10–15% | < 5ms | `rbac.cache.redis_ttl` |
| Database (Casbin) | < 5% miss | < 20ms | — |
| Middleware overhead | — | 1–6ms total | — |

Cache is automatically invalidated when policies or role assignments change, via RabbitMQ events published by the watcher.

To force a reload without restarting the server:
```bash
curl -X POST http://localhost:8080/ichi-go/api/v1/rbac/policies/reload \
  -H "Authorization: Bearer $TOKEN"
```

---

## REST API Reference

All RBAC endpoints require authentication (`Authorization: Bearer <token>`).

The app name is configured at `app.name` in `config.local.yaml` (default: `ichi-go`).

### Enforcement (`/{app}/api/enforce`)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/enforce/check` | Check a single permission |
| `POST` | `/enforce/batch` | Check multiple permissions at once |
| `GET` | `/enforce/my-permissions` | Get all permissions for the current user |

**Check permission request:**
```json
{
  "tenant_id": "acme-corp",
  "resource": "products",
  "action": "edit"
}
```

### Policy Management (`/{app}/api/v1/rbac/policies`)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/policies` | List all policies |
| `POST` | `/policies` | Add a policy |
| `DELETE` | `/policies` | Remove a policy |
| `GET` | `/policies/count` | Count policies |
| `POST` | `/policies/reload` | Reload policies from database |

### Role Management (`/{app}/api/v1/rbac/roles`)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/roles` | List roles (paginated) |
| `POST` | `/roles` | Create a role |
| `GET` | `/roles/:id` | Get a role |
| `PUT` | `/roles/:id` | Update a role |
| `DELETE` | `/roles/:id` | Delete a role |
| `GET` | `/roles/:id/permissions` | Get role with all permissions |
| `GET` | `/roles/:roleId/users` | List users with this role |

### User-Role Management (`/{app}/api/v1/rbac/users`)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/users/:userId/roles` | Get all roles for a user |
| `GET` | `/users/:userId/roles/active` | Get active (non-expired) roles |
| `POST` | `/users/:userId/roles` | Assign a role to a user |
| `DELETE` | `/users/:userId/roles/:roleSlug` | Revoke a role from a user |

**Assign role request:**
```json
{
  "role_slug": "tenant-editor",
  "tenant_id": "acme-corp",
  "reason": "Promoted to editor"
}
```

### Audit Logs (`/{app}/api/v1/rbac/audit`)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/audit/logs` | Query logs with filters |
| `GET` | `/audit/stats` | Aggregated statistics |
| `POST` | `/audit/export` | Export to CSV or JSON |
| `GET` | `/audit/mutations` | Recent policy/role changes |
| `GET` | `/audit/decisions` | Recent permission decisions |

**Query logs with filters:**
```
GET /audit/logs?tenant_id=acme-corp&action=role_assigned&start_date=2026-01-01&limit=100
```

---

## Audit Logging

**What is logged automatically:**

| Event | Logged by default |
|-------|------------------|
| Role assigned / revoked | Yes (`log_mutations: true`) |
| Policy added / removed | Yes (`log_mutations: true`) |
| Permission check decisions | No (`log_decisions: false`) — enable only for debugging |

**Compliance features:**

- **Retention**: Configure `audit.retention_days` (default: 2555 = 7 years, SOC2 requirement)
- **PII anonymization**: Email addresses are hashed with SHA-256 when `anonymize_pii: true` (GDPR)
- **Async logging**: Audit writes are non-blocking and do not add latency to permission checks

---

## Database Tables

Six tables are created by `db/migrations/schema/20260117_001_create_rbac_casbin_tables.sql`:

| Table | Description |
|-------|-------------|
| `casbin_rule` | Casbin policy rules (`v0`=role, `v1`=tenant, `v2`=resource, `v3`=action) |
| `rbac_roles` | Role definitions (name, slug, description, tenant scope) |
| `rbac_permissions` | 146 pre-defined permissions |
| `rbac_user_roles` | User-role assignments (with optional expiry) |
| `rbac_audit_log` | Full audit trail for SOC2/GDPR compliance |
| `rbac_platform_permissions` | Platform-level global permissions |

Run migrations and seeds:
```bash
make migration-up
make seed-run
```

---

## Troubleshooting

| Symptom | Diagnosis | Fix |
|---------|-----------|-----|
| 403 on every request | No tenant context | Add `X-Tenant-Id: <tenant>` header, or configure `default_tenant` in config |
| Permission denied (unexpected) | Stale cache | `POST /{app}/api/v1/rbac/policies/reload` |
| Very slow permission checks | Cache not hitting | Check Redis connection; verify `rbac.cache.enabled: true` |
| User has role but still denied | Wrong tenant or role not seeded | `GET /{app}/api/v1/rbac/users/{id}/roles` to inspect |
| `No tenant ID in context` error | TenantContextMiddleware not applied | Register `TenantContextMiddleware` before `RBACEnforcementMiddleware` |
