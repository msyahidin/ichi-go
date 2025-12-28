# Multi-Strategy Versioning - Quick Start

## What's New

The versioning package now supports **multiple versioning strategies**:

✅ **Semantic** - `v1`, `v2`, `v3` (default)  
✅ **Date Monthly** - `2026-01`, `2026-12`  
✅ **Date Daily** - `20260115`, `20261231`  
✅ **Custom** - Define your own pattern

## Quick Examples

### Example 1: Date-Based Monthly (Like Your Request)

```yaml
# config.yaml
api:
  versioning:
    enabled: true
    strategy: "date"
    default_version: "2026-01"
    supported_versions: ["2026-01"]
```

**Result:**
```bash
POST /api/2026-01/auth/login  # Instead of /api/v1/auth/login
```

### Example 2: Semantic (Default)

```yaml
# config.yaml
api:
  versioning:
    enabled: true
    strategy: "semantic"
    default_version: "v1"
    supported_versions: ["v1"]
```

**Result:**
```bash
POST /api/v1/auth/login
```

### Example 3: Custom Quarterly

```yaml
# config.yaml
api:
  versioning:
    enabled: true
    strategy: "custom"
    default_version: "release-2026-Q1"
    custom_pattern: "^release-\\d{4}-Q[1-4]$"
    supported_versions: ["release-2026-Q1"]
```

**Result:**
```bash
POST /api/release-2026-Q1/auth/login
```

## Implementation (3 Steps)

### Step 1: Choose Strategy in Config

```yaml
api:
  versioning:
    enabled: true
    strategy: "date"  # or "semantic", "date_daily", "custom"
    default_version: "2026-01"
    supported_versions: ["2026-01"]
```

### Step 2: Initialize Strategy in Code

```go
// In middlewares/registry.go
func Init(e *echo.Echo, mainConfig *config.Config) {
    // ... other middleware ...
    
    versionConfig := mainConfig.Versioning()
    if versionConfig != nil && versionConfig.Enabled {
        // Initialize the strategy based on config
        if err := versionConfig.InitializeStrategy(); err != nil {
            logger.Fatalf("Failed to initialize version strategy: %v", err)
        }
        
        e.Use(VersionLogger())
        e.Use(VersionValidator(versionConfig))
        e.Use(VersionDeprecation())
    }
}
```

### Step 3: Use in Routes (Same for All Strategies!)

```go
// Your route code stays the same regardless of strategy
func (c *AuthController) RegisterRoutesV1(...) {
    version := versioning.GetCurrentDateVersion() // For date-based
    // OR
    version := versioning.V1 // For semantic
    
    vr := versioning.NewVersionedRoute(serviceName, version, Domain)
    publicGroup := vr.Group(e)
    publicGroup.POST("/login", c.Login)
}
```

## Helper Functions for Date-Based

```go
// Monthly
current := versioning.GetCurrentDateVersion()              // "2026-01"
custom := versioning.NewDateVersion(2026, 1)               // "2026-01"

// Daily
current := versioning.GetCurrentDateDailyVersion()         // "20260115"
custom := versioning.NewDateDailyVersion(2026, 1, 15)      // "20260115"

// Parse
time, _ := versioning.ParseDateVersion("2026-01")          // time.Time
time, _ := versioning.ParseDateDailyVersion("20260115")    // time.Time
```

## Complete Example: Monthly Date-Based

```yaml
# config.yaml
app:
  name: "ichi-go"

api:
  versioning:
    enabled: true
    strategy: "date"
    default_version: "2026-01"
    supported_versions:
      - "2025-12"  # Last month (still supported)
      - "2026-01"  # This month (current)
    deprecation:
      header_enabled: true
      sunset_notification_days: 90
```

```go
// controller/auth_routes.go
package auth

import (
    "ichi-go/pkg/versioning"
    "github.com/labstack/echo/v4"
)

func (c *AuthController) RegisterRoutes(serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
    // Current month
    currentMonth := versioning.GetCurrentDateVersion()
    c.registerForVersion(serviceName, e, auth, currentMonth)
    
    // Also support previous month during transition
    // lastMonth := versioning.NewDateVersion(2025, 12)
    // c.registerForVersion(serviceName, e, auth, lastMonth)
}

func (c *AuthController) registerForVersion(serviceName string, e *echo.Echo, auth *authenticator.Authenticator, version versioning.APIVersion) {
    vr := versioning.NewVersionedRoute(serviceName, version, "auth")
    
    publicGroup := vr.Group(e)
    publicGroup.POST("/login", c.Login)
    publicGroup.POST("/register", c.Register)
    
    protectedGroup := vr.Group(e)
    protectedGroup.Use(auth.AuthenticateMiddleware())
    protectedGroup.GET("/me", c.Me)
}
```

**Result:**
```bash
# Endpoints created
POST /ichi-go/api/2026-01/auth/login
POST /ichi-go/api/2026-01/auth/register
GET  /ichi-go/api/2026-01/auth/me
```

## Strategy Comparison

| Strategy | Format | Example | Best For |
|----------|--------|---------|----------|
| Semantic | `vN` | `v1`, `v2` | Major releases |
| Date Monthly | `YYYY-MM` | `2026-01` | Monthly releases |
| Date Daily | `YYYYMMDD` | `20260115` | Daily deployments |
| Custom | Your choice | `release-2026-Q1` | Special needs |

## Key Benefits

1. **Same Code** - Route registration code works with any strategy
2. **Config-Driven** - Change strategy without code changes
3. **Automatic Validation** - Each strategy validates format automatically
4. **Helper Functions** - Built-in helpers for date-based versions
5. **Backward Compatible** - Existing semantic code still works

## Documentation

- **MULTI_STRATEGY_GUIDE.md** - Complete guide with all strategies
- **STRATEGY_CONFIGS.md** - Ready-to-use config examples
- **API_VERSIONING_STRATEGY.md** - Original comprehensive guide
- **MIGRATION_GUIDE.md** - Team migration instructions

## Testing

```bash
# Run tests
cd pkg/versioning
go test -v

# Should see tests for all strategies:
# ✓ TestSemanticVersionStrategy
# ✓ TestDateVersionStrategy  
# ✓ TestDateDailyVersionStrategy
# ✓ TestCustomVersionStrategy
```

## Common Patterns

### Auto-Update for Date-Based

```go
// Update version monthly via cron/deployment
func updateToCurrentMonth() {
    current := versioning.GetCurrentDateVersion()
    // Update config with new default_version
}
```

### Support Multiple Concurrent Versions

```go
// Support last 3 months
func RegisterRoutes(serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
    now := time.Now()
    
    for i := 0; i < 3; i++ {
        month := now.AddDate(0, -i, 0)
        version := versioning.NewDateVersion(month.Year(), int(month.Month()))
        c.registerForVersion(serviceName, e, auth, version)
    }
}
```

## Migration from v1/v2 to Date-Based

See `MULTI_STRATEGY_GUIDE.md` section "Migration Between Strategies"

Quick version:
1. Change `strategy: "date"` in config
2. Update route registration to use date versions
3. Optionally support both during transition

## Next Steps

1. Choose your strategy (see STRATEGY_CONFIGS.md)
2. Update config.yaml
3. Initialize strategy in middleware
4. Update routes (optional - same code works)
5. Test and deploy

That's it! The versioning package handles the rest.