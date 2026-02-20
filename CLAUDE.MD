# ichi-go - Project Guide for Claude Code

This document provides Claude Code with essential context about the ichi-go project architecture, patterns, and development workflows.

## Project Overview

**ichi-go** is a production-ready Go backend template built on clean architecture principles. It provides a comprehensive foundation for developing scalable REST API services with enterprise-grade features including JWT authentication, message queuing, caching, and comprehensive validation.

**Tech Stack:**
- Go 1.24+
- Echo v4 (HTTP framework)
- Bun ORM (MySQL dialect)
- Redis (caching with LZ4 compression)
- RabbitMQ (message queue)
- samber/do (dependency injection)
- Viper (configuration management)
- Goose (SQL migrations)

## Architecture Patterns

### Clean Architecture Structure

The project follows domain-driven design with clear separation of concerns:

```
internal/
├── applications/          # Domain applications (auth, user, health, notification, rbac)
│   └── {domain}/
│       ├── controllers/   # HTTP handlers (presentation layer)
│       ├── services/      # Business logic
│       ├── repositories/  # Data access
│       ├── dto/          # Request/response models
│       ├── validators/   # Domain-specific validation
│       ├── models/       # Domain entities
│       └── register.go   # DI setup with samber/do
├── infra/                # Infrastructure layer
│   ├── database/        # Bun ORM setup
│   ├── cache/           # Redis client
│   ├── queue/           # RabbitMQ integration
│   └── authz/           # Casbin RBAC enforcer, adapter, cache, watcher
└── middlewares/         # HTTP middlewares
```

### Dependency Injection Pattern (samber/do)

The project uses samber/do for runtime dependency injection:

```go
// In domain's register.go
func Register(injector do.Injector, serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
    // Register repository
    do.ProvideNamed(injector, serviceName+".user.repository",
        func(i do.Injector) (*repositories.UserRepository, error) {
            db := do.MustInvoke[*database.DB](i)
            return repositories.NewUserRepository(db), nil
        })

    // Register service with dependency on repository
    do.ProvideNamed(injector, serviceName+".user.service",
        func(i do.Injector) (*services.UserService, error) {
            repo := do.MustInvokeNamed[*repositories.UserRepository](
                i, serviceName+".user.repository")
            return services.NewUserService(repo), nil
        })
}
```

**Key Points:**
- Always use named dependencies with `serviceName` prefix for domain isolation
- Use `MustInvoke` for required dependencies
- Use `MustInvokeNamed` when retrieving named dependencies
- Cleanup is handled automatically via `injector.Shutdown()` in main.go

## Key Features & Implementation Patterns

### 1. JWT Authentication

Location: `pkg/authenticator/`

**Configuration:**
- Supports multiple algorithms: HS256/384/512, RS256/384/512, ES256/384/512
- Configurable via `config.*.yaml` under `auth.jwt`
- Skip paths configured in `auth.jwt.skip_paths`

**Usage in handlers:**
```go
// Get user ID from context
userID := requestctx.GetUserID(ctx.Request().Context())

// Get full request context
rc := requestctx.FromRequest(ctx.Request())
```

### 2. Validation System

Location: `pkg/validator/`

**Features:**
- Multilingual support (English/Indonesian)
- Domain-isolated validators
- Language detection via `X-Language` header or `Accept-Language`

**Pattern for new validators:**
```go
// In domain's validators/ directory
type CustomValidator struct {
    Tag: "custom_tag",
    Fn: func(fl validator.FieldLevel) bool {
        // Validation logic
        return true
    },
    RegisterTrans: func(trans ut.Translator) error {
        return trans.Add("custom_tag", "Error message {0}", true)
    },
}
```

**Usage in controllers:**
```go
if err := validator.BindAndValidate(c, &req); err != nil {
    return response.Error(c, http.StatusBadRequest, err)
}
```

### 3. Message Queue (RabbitMQ)

Location: `internal/infra/queue/`

**Consumer Registration:**
```go
// In internal/infra/queue/registry.go
func init() {
    RegisterConsumer(ConsumerRegistration{
        Name: "handler_name",
        Description: "Handler description",
        ConsumeFunc: func(ctx context.Context, msg rabbitmq.Message) error {
            // Process message
            return nil
        },
    })
}
```

**Publishing Messages:**
```go
producer.Publish(ctx, rabbitmq.PublishParams{
    RoutingKey: "event.name",
    Body: data,
})
```

**Important Notes:**
- Always handle context cancellation for graceful shutdown
- Return error for retryable failures, nil for permanent failures
- Configure consumers in `config.*.yaml` under `queue.rabbitmq.consumers`

### 4. Database & Bun ORM

Location: `internal/infra/database/`

**Base Repository Pattern:**
```go
type UserRepository struct {
    *bun.BaseRepository[models.User]
}

func NewUserRepository(db *database.DB) *UserRepository {
    return &UserRepository{
        BaseRepository: bun.NewBaseRepository[models.User](db),
    }
}
```

**Base Repository provides:**
- `FindByID(ctx, id)` - Find single record
- `Create(ctx, model)` - Insert record
- `Update(ctx, id, model)` - Update record
- `Delete(ctx, id)` - Delete record
- `FindAll(ctx, query)` - Find with pagination

**Migration Management:**
- Schema migrations: `db/migrations/schema/`
- Data migrations: `db/migrations/data/`
- Seeds: `db/migrations/seeds/`
- See Makefile for all migration commands

### 5. Caching (Redis)

Location: `internal/infra/cache/`

**Features:**
- Automatic LZ4 compression
- TTL support
- Generic interface

**Usage:**
```go
cache := do.MustInvoke[*cache.Cache](injector)

// Set with TTL
cache.Set(ctx, "key", value, 1*time.Hour)

// Get
var result Type
cache.Get(ctx, "key", &result)
```

### 6. API Versioning

The project supports date-based API versioning:

**Configuration:**
```yaml
api:
  versioning:
    enabled: true
    strategy: "date"
    default_version: "202601"
    supported_versions: ["202511", "202512", "202601"]
    deprecation:
      header_enabled: true
      sunset_notification_days: 90
```

**Usage in routes:**
```go
// Routes are versioned under /api/v{version}/
e.GET("/api/v1/users", controller.GetUsers)
```

## Development Workflows

### Creating New Domain Features

When adding a new domain feature (e.g., "product" domain):

1. **Use Code Generator:**
```bash
# Generate full CRUD module
go run ./pkg/ichigen/cmd/main.go g full product --domain=catalog --crud

# Or use Makefile
make ichigen-example-full
```

2. **Generated Structure:**
```
internal/applications/product/
├── controllers/
│   └── product_controller.go
├── services/
│   └── product_service.go
├── repositories/
│   └── product_repository.go
├── dto/
│   ├── create_product_dto.go
│   └── update_product_dto.go
├── validators/
│   └── product_validators.go
├── models/
│   └── product.go
└── register.go
```

3. **Register in main.go:**
```go
// Import the domain
import productapp "ichi-go/internal/applications/product"

// Register with DI
productapp.Register(injector, serviceName, e, auth)
```

### Database Migrations

```bash
# Create schema migration
make migration-create name=create_products_table

# Run migrations
make migration-up

# Rollback last migration
make migration-down

# Check status
make migration-status

# Create seed file
make seed-run
```

### Testing

**Unit Testing Pattern:**
```go
func TestService_Method(t *testing.T) {
    // Setup mocks
    mockRepo := new(mock_domain.Repository)
    service := services.NewService(mockRepo)

    // Define expectations
    mockRepo.On("Method", mock.Anything, mock.Anything).
        Return(expectedResult, nil)

    // Execute
    result, err := service.Method(ctx, input)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
    mockRepo.AssertExpectations(t)
}
```

**Generate mocks:**
```bash
mockery --all --dir internal/applications --output mocks --keeptree
```

### Configuration Management

**File Priority:**
1. `config.local.yaml` (local dev)
2. `config.dev.yaml` (dev environment)
3. `config.staging.yaml` (staging)
4. `config.prod.yaml` (production)

**Set environment:**
```bash
export APP_ENV=prod
go run cmd/main.go
```

## Common Patterns & Best Practices

### Error Handling

```go
// Define domain errors
var (
    ErrNotFound = errors.New("resource not found")
    ErrInvalidInput = errors.New("invalid input")
)

// In service layer
func (s *Service) GetByID(ctx context.Context, id string) (*Model, error) {
    model, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to get resource: %w", err)
    }
    return model, nil
}
```

### Response Helpers

Location: `pkg/utils/response/`

```go
// Success response
return response.Success(c, data)

// Error response
return response.Error(c, http.StatusBadRequest, err)

// Paginated response
return response.SuccessWithPagination(c, data, pagination)
```

### Request Context

```go
// In middleware or controller
rc := requestctx.FromRequest(c.Request())

// Access user ID
userID := requestctx.GetUserID(ctx)

// Access request ID
requestID := rc.RequestID

// Access language
lang := rc.Language
```

### Logging

Location: `pkg/logger/`

```go
// Standard logging
logger.Info("Message")
logger.Debugf("Debug message: %v", data)
logger.Errorf("Error: %v", err)

// With context (includes request ID)
logger.WithContext(ctx).Info("Processing request")
```

## Important Files & Locations

### Configuration
- `config/config.go` - Main config loader
- `config.example.yaml` - Example configuration
- `config.local.yaml` - Local development config (gitignored)

### Entry Points
- `cmd/main.go` - Application entry point
- `cmd/server/rest_server.go` - HTTP server setup
- `cmd/server/queue_server.go` - RabbitMQ workers
- `cmd/server/web_server.go` - Web/template routes

### Infrastructure
- `internal/infra/database/database.go` - Database connection
- `internal/infra/cache/cache.go` - Redis connection
- `internal/infra/queue/rabbitmq/` - RabbitMQ implementation

### Utilities
- `pkg/authenticator/` - JWT implementation
- `pkg/validator/` - Validation framework
- `pkg/utils/response/` - HTTP response builders
- `pkg/utils/dto/` - DTO mapping utilities
- `pkg/requestctx/` - Request context management

### Code Generation
- `pkg/ichigen/` - Code generator CLI
- `Makefile` - Build automation and commands

### Documentation
- `docs/STARTER_GUIDE.md` - New developer onboarding (Day 1 guide)
- `docs/RBAC.md` - RBAC system architecture and usage guide
- `docs/ADDING_NEW_SERVICE.md` - Step-by-step guide for adding a new domain
- `docs/testing/` - Testing guides (TESTING.md, TESTING_QUICKSTART.md, TESTING_CHEATSHEET.md)

## Common Makefile Commands

```bash
# Development
air                          # Hot reload
make ichigen-install         # Install code generator globally

# Database
make migration-create name=  # Create migration
make migration-up            # Run migrations
make migration-down          # Rollback last
make migration-status        # Check status
make seed-run               # Run all seeds
make db-reset-dev           # Reset + migrate + seed

# Code Generation
make ichigen-example-full    # Generate full CRUD example

# Testing
go test ./...               # Run all tests
go test -cover ./...        # With coverage
```

## Security Considerations

1. **JWT Keys**: Change default secret keys in production
2. **CORS**: Configure `http.cors.allow_origins` appropriately
3. **Database**: Use strong credentials, limit connections
4. **Validation**: Always validate input at API boundaries
5. **Error Messages**: Don't expose sensitive info in error responses
6. **Queue**: Validate message payloads before processing

## Performance Tips

1. **Caching**: Use Redis for frequently accessed data
2. **Connection Pools**: Configured in config.yaml
3. **Pagination**: Always paginate list endpoints
4. **Indexes**: Add database indexes for queried fields
5. **Queue Workers**: Adjust `worker_pool_size` based on load

## Troubleshooting

### Database Connection Issues
- Check `database` config in config.yaml
- Verify MySQL is running: `mysql -u root -p`
- Run migrations: `make migration-up`

### Redis Connection Issues
- Check `cache` config in config.yaml
- Verify Redis is running: `redis-cli ping`

### RabbitMQ Issues
- Check `queue.rabbitmq` config
- Verify RabbitMQ is running: `rabbitmqctl status`
- Check exchange setup in config.yaml

### Validation Not Working
- Check validator registration in domain's validators/
- Verify language configuration
- Check translation files

## When Modifying Code

### Adding New Domain
1. Use ichigen generator: `ichigen g full {name} --domain={domain}`
2. Create migrations for database schema
3. Register domain in `cmd/main.go`
4. Add tests for new functionality

### Adding New Endpoint
1. Add method to controller
2. Add route in domain's `register.go`
3. Add validation in DTO
4. Update Swagger docs if enabled

### Adding New Queue Consumer
1. Register in `internal/infra/queue/registry.go`
2. Add consumer config in `config.*.yaml`
3. Implement handler function with context support
4. Test graceful shutdown

### Modifying Database Schema
1. Create migration: `make migration-create name=`
2. Write up and down migrations
3. Test migration: `make migration-up`
4. Update model structs
5. Update repository if needed

## RBAC System (Role-Based Access Control)

Location: `internal/applications/rbac/`, `internal/infra/authz/`

### Overview

The ichi-go RBAC system provides enterprise-grade authorization with:
- **Multi-tenant support** with tenant isolation
- **Three-layer permission model** (Platform → Application → Resource)
- **Multi-tier caching** (L1 memory + L2 Redis) for performance
- **Event-driven cache invalidation** via RabbitMQ
- **Comprehensive audit logging** for SOC2/GDPR compliance
- **Casbin v3 integration** for flexible policy enforcement

### Architecture

**Three-Layer Authorization Model:**

1. **Layer 1 - Platform Permissions** (Global)
   - Platform admin access
   - Cross-tenant operations
   - System maintenance
   - Example: `platform:admin:access`

2. **Layer 2 - Application Permissions** (Tenant-scoped)
   - Standard RBAC via Casbin
   - Tenant-isolated policies
   - Role-based permissions
   - Example: `users:edit` in tenant "acme"

3. **Layer 3 - Resource Permissions** (Future ABAC)
   - Attribute-based access control
   - Fine-grained resource permissions
   - Example: `product:123:edit`

### Key Components

**Infrastructure** (`internal/infra/authz/`)
- `enforcer/` - Casbin v3 enforcer wrapper with caching
- `adapter/` - Bun ORM adapter for Casbin
- `cache/` - Multi-tier cache (memory + Redis)
- `watcher/` - RabbitMQ event watcher for cache invalidation
- `circuit_breaker/` - Resilience against DB failures

**Domain** (`internal/applications/rbac/`)
- `models/` - 7 domain entities (Role, Permission, UserRole, AuditLog, etc.)
- `repositories/` - 6 data access layers
- `services/` - 5 business logic services
- `controllers/` - 5 REST API controllers (36 endpoints)
- `dto/` - Request/response models
- `constants/` - Error codes and permission constants

**Middlewares** (`internal/middlewares/`)
- `tenant_context_middleware.go` - Multi-tenant context resolution
- `rbac_enforcement_middleware.go` - Permission checking
- `rbac_audit_middleware.go` - Comprehensive audit logging

### Configuration

Add to `config.yaml`:

```yaml
rbac:
  mode: "hybrid"  # or "multi-tenant", "single-tenant"
  default_tenant: "default"
  model_path: "./config/rbac_model.conf"

  performance:
    max_policies_per_tenant: 10000
    policy_load_batch_size: 1000

  cache:
    enabled: true
    ttl_seconds: 300
    max_size: 10000
    compression: true
    l1_enabled: true
    l1_max_size: 1000
    l1_ttl_seconds: 60

  audit:
    enabled: true
    retention_days: 2555  # 7 years for compliance
    async_logging: true

  features:
    platform_permissions: true
    resource_permissions: false
    permission_groups: true
```

### Usage Examples

**1. Check Permission (Service Layer)**

```go
// Inject enforcement service
enforcementService := do.MustInvoke[*services.EnforcementService](injector)

// Check single permission
allowed, err := enforcementService.CheckPermission(
    ctx,
    userID,      // int64
    tenantID,    // string
    "products",  // resource
    "edit",      // action
)

// Check multiple permissions (batch)
checks := []services.PermissionCheck{
    {Resource: "products", Action: "view"},
    {Resource: "orders", Action: "view"},
}
results, err := enforcementService.CheckBatch(ctx, userID, tenantID, checks)

// Get all user permissions
permissions, err := enforcementService.GetUserPermissions(ctx, userID, tenantID)
// Returns: ["products:view", "products:edit", "orders:view", ...]
```

**2. Protect Routes (Middleware)**

```go
// Global RBAC enforcement (auto-maps HTTP methods to actions)
enforcementService := do.MustInvoke[*services.EnforcementService](injector)
e.Use(RBACEnforcementMiddleware(enforcementService, DefaultRBACConfig()))

// Protect specific routes
admin := e.Group("/admin")
admin.Use(RequirePermission(enforcementService, "admin", "access"))

// In handlers
func myHandler(c echo.Context) error {
    ctx := c.Request().Context()
    userID := requestctx.GetUserIDAsInt64(ctx)
    tenantID := requestctx.GetTenantID(ctx)

    // Manual permission check
    allowed, _ := enforcementService.CheckPermission(
        ctx, userID, tenantID, "special-resource", "access")

    if !allowed {
        return echo.NewHTTPError(http.StatusForbidden, "Access denied")
    }

    return c.JSON(http.StatusOK, "Success")
}
```

**3. Tenant Context Resolution**

```go
// Auto-detect tenant from multiple sources
e.Use(TenantContextMiddleware(TenantConfig{
    Strategy: "auto", // Tries header → subdomain → path → query
    RequireTenant: true,
}))

// Use specific strategy
e.Use(TenantContextMiddleware(TenantConfig{
    Strategy: "header",
    HeaderName: "X-Tenant-Id",
}))

// In handlers
func myHandler(c echo.Context) error {
    tenantID := requestctx.GetTenantID(c.Request().Context())
    // Use tenant ID
}
```

**4. Role Management**

```go
roleService := do.MustInvoke[*services.RoleService](injector)

// Create role
role := &models.Role{
    Name: "Product Manager",
    Slug: "product-manager",
    TenantID: &tenantID,
}
err := roleService.CreateRole(ctx, role)

// Assign role to user
userRoleService := do.MustInvoke[*services.UserRoleService](injector)
err = userRoleService.AssignRole(
    ctx,
    userID,
    "product-manager",
    tenantID,
    adminID,  // assigned by
    "Promoted to product manager",
)
```

**5. Audit Log Querying**

```go
auditService := do.MustInvoke[*services.AuditService](injector)

// Query audit logs
query := repositories.AuditQuery{
    TenantID: &tenantID,
    StartDate: &startDate,
    EndDate: &endDate,
    Action: stringPtr("role_assigned"),
    Limit: 100,
}
logs, total, err := auditService.QueryAuditLogs(ctx, query)

// Export for compliance
err = auditService.ExportToCSV(ctx, query, "/tmp/audit-export.csv")
err = auditService.ExportToJSON(ctx, query, "/tmp/audit-export.json")

// Get statistics
stats, err := auditService.GetAuditStats(ctx, tenantID, startDate, endDate)
// Returns: map[action]count
```

### REST API Endpoints

**Permission Checking:**
- `POST /api/v1/rbac/enforce/check` - Check single permission
- `POST /api/v1/rbac/enforce/batch` - Check multiple permissions
- `GET /api/v1/rbac/enforce/my-permissions` - Get current user permissions

**Policy Management:**
- `GET /api/v1/rbac/policies` - List policies
- `POST /api/v1/rbac/policies` - Add policy
- `DELETE /api/v1/rbac/policies` - Remove policy
- `POST /api/v1/rbac/policies/reload` - Reload from database

**Role Management:**
- `GET /api/v1/rbac/roles` - List roles
- `POST /api/v1/rbac/roles` - Create role
- `GET /api/v1/rbac/roles/:id` - Get role
- `PUT /api/v1/rbac/roles/:id` - Update role
- `DELETE /api/v1/rbac/roles/:id` - Delete role

**User-Role Assignments:**
- `GET /api/v1/rbac/users/:userId/roles` - Get user roles
- `POST /api/v1/rbac/users/:userId/roles` - Assign role
- `DELETE /api/v1/rbac/users/:userId/roles/:roleSlug` - Revoke role

**Audit Logs:**
- `GET /api/v1/rbac/audit/logs` - Query audit logs
- `GET /api/v1/rbac/audit/stats` - Get statistics
- `POST /api/v1/rbac/audit/export` - Export logs (CSV/JSON)

### Database Schema

```sql
-- Core RBAC tables
casbin_rule                 -- Casbin policies (v0=role, v1=tenant, v2=resource, v3=action)
rbac_roles                  -- Application roles
rbac_permissions            -- Permission definitions
rbac_user_roles             -- User-role assignments
rbac_audit_log              -- Audit trail (SOC2/GDPR)
rbac_platform_permissions   -- Platform-level permissions
```

### Performance Characteristics

- **L1 Cache (Memory)**: 80-90% hit rate, < 1ms latency
- **L2 Cache (Redis)**: 10-15% hit rate, < 5ms latency
- **Database (Casbin)**: < 5% miss rate, < 20ms latency
- **Middleware Overhead**: ~1-6ms per request (with cache)

### Cache Invalidation

Events automatically published to RabbitMQ:
```go
// Policy changes
watcher.RBACEvent{
    Action: "policy_added",
    TenantID: "acme",
    Details: {ReloadPolicy: true},
}

// Role assignments
watcher.RBACEvent{
    Action: "role_assigned",
    TenantID: "acme",
    SubjectID: "user123",
    Details: {Role: "admin"},
}
```

### Troubleshooting

**No tenant context:**
- Check `X-Tenant-Id` header is sent
- Enable `TenantContextMiddleware`
- Set `default_tenant` in config

**Permission denied but should be allowed:**
- Check user has role assigned: `GET /api/v1/rbac/users/{id}/roles`
- Check role has permission: `GET /api/v1/rbac/roles/{id}/permissions`
- Check Casbin policy: `GET /api/v1/rbac/policies?role=admin&tenant_id=acme`
- Clear cache: `POST /api/v1/rbac/policies/reload`

**Slow permission checks:**
- Verify cache is enabled in config
- Check Redis connection
- Monitor cache hit ratio in logs
- Increase L1 cache size if needed

### Migration

Run RBAC migrations:
```bash
make migration-up
make seed-run
```

This creates:
- 10 RBAC tables
- 146 default permissions
- 9 system roles (super-admin to guest)
- Default Casbin policies

### System Roles

Pre-seeded roles:
- `super-admin` - Full system access (platform + all tenants)
- `platform-admin` - Platform-level operations
- `tenant-admin` - Full tenant access
- `tenant-editor` - Edit permissions in tenant
- `tenant-viewer` - Read-only in tenant
- `developer` - Development tools access
- `support` - Support operations
- `auditor` - Audit log access only
- `guest` - Minimal access

## References

- [Echo Framework](https://echo.labstack.com/)
- [Bun ORM](https://bun.uptrace.dev/)
- [samber/do](https://github.com/samber/do)
- [Go-Playground Validator](https://pkg.go.dev/github.com/go-playground/validator/v10)
- [RabbitMQ Tutorials](https://www.rabbitmq.com/tutorials)
- [Casbin v3 Documentation](https://casbin.org/docs/overview)

---

**Note for Claude Code**: When making changes, always:
- Follow the existing architectural patterns
- Use the code generator for new domains when appropriate
- Maintain domain isolation
- Add proper error handling and validation
- Write tests for new functionality
- Update migrations for schema changes
- Consider backward compatibility for API changes
