# Mockery Configuration Portability Guide

This guide explains how the `.mockery.yaml` configuration is designed to work with any project name.

## The Problem

Traditional mockery configurations use absolute package paths:

```yaml
# ❌ NOT PORTABLE - Only works with "ichi-go" module
packages:
  ichi-go/internal/applications/auth/service:
    interfaces:
      AuthService:
```

**Issues:**
- Doesn't work if you rename your module
- Can't be reused in other projects (notification-service, cart-service, etc.)
- Requires manual find-replace when forking/templating
- Module name hardcoded throughout config

## The Solution

Use **relative paths** with `./` prefix:

```yaml
# ✅ PORTABLE - Works with ANY module name
packages:
  ./internal/applications/auth/service:
    config:
      all: true
```

**Benefits:**
- Works with any module name (ichi-go, notification-service, cart-service)
- Configuration is copy-paste ready for new projects
- No need to modify when renaming module
- Template-friendly for microservices architecture

## How It Works

### 1. Relative Package Paths

```yaml
packages:
  # Relative path from project root
  ./internal/applications/auth/service:
    config:
      all: true  # Generate mocks for all exported interfaces

  ./internal/applications/user/repository:
    config:
      all: true

  ./pkg/authenticator:
    config:
      all: true
```

### 2. Dynamic Directory Placement

```yaml
# Global setting - applies to all packages
dir: "{{.InterfaceDir}}/mocks"
```

**Template Variables:**
- `{{.InterfaceDir}}` - Directory where the interface is defined
- Automatically creates `mocks/` subdirectory next to source

**Example:**
```
Source:  internal/applications/auth/service/auth_service.go
Mock:    internal/applications/auth/service/mocks/mock_auth_service.go
```

### 3. Mock Naming

```yaml
mockname: "Mock{{.InterfaceName}}"
filename: "mock_{{.InterfaceName | snakecase}}.go"
outpkg: "mocks"
```

**Results:**
- Interface: `AuthService`
- Mock name: `MockAuthService`
- File: `mock_auth_service.go`
- Package: `mocks`

## Complete Configuration

```yaml
# Mockery Configuration
# Works with any project name by using relative paths

# Global settings
quiet: false
with-expecter: true
all: true
inpackage: false
testonly: false
exported: true
keeptree: true

# Dynamic directory - mocks placed next to source
dir: "{{.InterfaceDir}}/mocks"

# Naming patterns
mockname: "Mock{{.InterfaceName}}"
filename: "mock_{{.InterfaceName | snakecase}}.go"
outpkg: "mocks"

# Packages with relative paths
packages:
  ./internal/applications/auth/service:
    config:
      all: true

  ./internal/applications/user/repository:
    config:
      all: true

  ./pkg/authenticator:
    config:
      all: true
```

## Usage Examples

### Example 1: Works with ichi-go

```bash
# go.mod
module ichi-go

# .mockery.yaml
packages:
  ./internal/applications/auth/service:
    config:
      all: true

# Generated mock
internal/applications/auth/service/mocks/mock_auth_service.go

# Import in tests
import "ichi-go/internal/applications/auth/service/mocks"
```

### Example 2: Works with notification-service

```bash
# go.mod
module notification-service

# .mockery.yaml (SAME CONFIG - no changes needed!)
packages:
  ./internal/applications/auth/service:
    config:
      all: true

# Generated mock (SAME STRUCTURE)
internal/applications/auth/service/mocks/mock_auth_service.go

# Import in tests (different module name, same path)
import "notification-service/internal/applications/auth/service/mocks"
```

### Example 3: Works with cart-service

```bash
# go.mod
module github.com/mycompany/cart-service

# .mockery.yaml (STILL THE SAME!)
packages:
  ./internal/applications/auth/service:
    config:
      all: true

# Generated mock (SAME STRUCTURE)
internal/applications/auth/service/mocks/mock_auth_service.go

# Import in tests
import "github.com/mycompany/cart-service/internal/applications/auth/service/mocks"
```

## Adding New Packages

### For a New Domain

```yaml
packages:
  # ... existing packages ...

  # Add new notification service
  ./internal/applications/notification/service:
    config:
      all: true

  ./internal/applications/notification/repository:
    config:
      all: true
```

### For Specific Interfaces Only

```yaml
packages:
  ./internal/applications/payment/service:
    interfaces:
      PaymentService:       # Only mock this
      RefundService:        # And this
      # StripeClient won't be mocked
```

### For Infrastructure Components

```yaml
packages:
  ./internal/infra/email:
    config:
      all: true

  ./internal/infra/sms:
    config:
      all: true
```

## Migration from Absolute Paths

If you have an existing `.mockery.yaml` with absolute paths:

### Before (Not Portable)

```yaml
packages:
  ichi-go/internal/applications/auth/service:
    interfaces:
      AuthService:
        config:
          dir: "internal/applications/auth/service/mocks"

  ichi-go/internal/applications/user/repository:
    interfaces:
      UserRepository:
        config:
          dir: "internal/applications/user/repository/mocks"
```

### After (Portable)

```yaml
packages:
  ./internal/applications/auth/service:
    config:
      all: true  # Simpler - generates all interfaces

  ./internal/applications/user/repository:
    config:
      all: true
```

**Changes:**
1. Replace `module-name/` with `./`
2. Remove explicit `dir` configs (use global `{{.InterfaceDir}}/mocks`)
3. Use `all: true` instead of listing each interface
4. Set global `keeptree: true` for consistent structure

### Migration Steps

1. **Backup existing mocks:**
   ```bash
   make test-clean-mocks  # Or manually backup
   ```

2. **Update .mockery.yaml:**
   - Replace absolute paths with relative (`./`)
   - Simplify with `all: true`
   - Use global `dir: "{{.InterfaceDir}}/mocks"`

3. **Regenerate all mocks:**
   ```bash
   make test-generate-mocks
   ```

4. **Verify imports still work:**
   ```bash
   make test
   ```

## Best Practices

### ✅ DO

```yaml
# Use relative paths
./internal/applications/auth/service:
  config:
    all: true

# Use templates for directories
dir: "{{.InterfaceDir}}/mocks"

# Generate all interfaces by default
all: true
```

### ❌ DON'T

```yaml
# Don't use absolute module paths
ichi-go/internal/applications/auth/service:  # ❌

# Don't hardcode directories
config:
  dir: "internal/applications/auth/service/mocks"  # ❌

# Don't list every interface manually (unless needed)
interfaces:
  Interface1:
  Interface2:
  Interface3:  # ❌ Use all: true instead
```

## Troubleshooting

### Mocks not generating?

```bash
# Clean and regenerate
make test-clean-mocks
make test-generate-mocks

# Or manually
find . -type d -name "mocks" -exec rm -rf {} +
mockery
```

### Wrong import paths in tests?

The import path in your tests should match your `go.mod` module name:

```go
// If go.mod says: module notification-service
import "notification-service/internal/applications/auth/service/mocks"

// If go.mod says: module github.com/company/cart
import "github.com/company/cart/internal/applications/auth/service/mocks"
```

The `.mockery.yaml` config doesn't change - only the import in tests!

### Mocks in wrong directory?

Check your global `dir` setting:

```yaml
# Should be:
dir: "{{.InterfaceDir}}/mocks"

# Not:
dir: "mocks"  # ❌ Puts all mocks in project root
```

## Templates & Microservices

### Use as Project Template

This configuration is perfect for project templates:

1. Create template with `.mockery.yaml`
2. Developers clone and rename module in `go.mod`
3. Configuration still works - no changes needed!

```bash
# Developer workflow
git clone template-repo my-new-service
cd my-new-service

# Change module name
sed -i 's/template/my-new-service/' go.mod

# Generate mocks - just works!
make test-generate-mocks
```

### Microservices Architecture

Same `.mockery.yaml` across all services:

```
microservices/
├── auth-service/
│   ├── go.mod (module: auth-service)
│   └── .mockery.yaml (same config)
├── notification-service/
│   ├── go.mod (module: notification-service)
│   └── .mockery.yaml (same config)
├── cart-service/
│   ├── go.mod (module: cart-service)
│   └── .mockery.yaml (same config)
└── payment-service/
    ├── go.mod (module: payment-service)
    └── .mockery.yaml (same config)
```

All services use identical `.mockery.yaml` - copy-paste friendly!

## Summary

**Key Points:**
1. Use `./` for relative paths
2. Use `{{.InterfaceDir}}` template for dynamic directories
3. Set `all: true` to generate all interfaces
4. Configuration is module-name agnostic
5. Works perfectly with microservices and templates

**Benefits:**
- ✅ Copy-paste ready for new projects
- ✅ Works with any module name
- ✅ No manual modifications needed
- ✅ Template and microservice friendly
- ✅ Future-proof configuration

---

**Remember:** The `.mockery.yaml` config never changes. Only the import paths in your test files reflect the module name from `go.mod`!
