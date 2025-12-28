# Multi-Strategy Versioning Guide

## Overview

The versioning package now supports multiple versioning strategies:

1. **Semantic Versioning** - `v1`, `v2`, `v3` (default)
2. **Date-Based Monthly** - `2026-01`, `2026-12`
3. **Date-Based Daily** - `20260115`, `20261231`
4. **Custom Strategies** - Define your own pattern

## Quick Examples

### Semantic Versioning (Default)

```yaml
# config.yaml
api:
  versioning:
    enabled: true
    strategy: "semantic"
    default_version: "v1"
    supported_versions: ["v1", "v2"]
```

```bash
# Endpoints
POST /ichi-go/api/v1/auth/login
POST /ichi-go/api/v2/auth/login
```

### Date-Based Monthly Versioning

```yaml
# config.yaml
api:
  versioning:
    enabled: true
    strategy: "date"
    default_version: "2026-01"
    supported_versions: ["2025-12", "2026-01", "2026-02"]
```

```bash
# Endpoints
POST /ichi-go/api/2026-01/auth/login
POST /ichi-go/api/2025-12/auth/login
```

### Date-Based Daily Versioning

```yaml
# config.yaml
api:
  versioning:
    enabled: true
    strategy: "date_daily"
    default_version: "20260115"
    supported_versions: ["20260101", "20260115", "20260201"]
```

```bash
# Endpoints
POST /ichi-go/api/20260115/auth/login
POST /ichi-go/api/20260101/auth/login
```

### Custom Versioning

```yaml
# config.yaml
api:
  versioning:
    enabled: true
    strategy: "custom"
    default_version: "release-2026-Q1"
    custom_pattern: "^release-\\d{4}-Q[1-4]$"
    supported_versions: 
      - "release-2025-Q4"
      - "release-2026-Q1"
      - "release-2026-Q2"
```

```bash
# Endpoints
POST /ichi-go/api/release-2026-Q1/auth/login
POST /ichi-go/api/release-2025-Q4/auth/login
```

## Implementation

### Step 1: Choose Your Strategy

Update `config.yaml`:

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
// In your main initialization (e.g., middlewares/registry.go)
func Init(e *echo.Echo, mainConfig *config.Config) {
    // ... other middleware ...
    
    versionConfig := mainConfig.Versioning()
    if versionConfig != nil && versionConfig.Enabled {
        // Initialize the strategy
        if err := versionConfig.InitializeStrategy(); err != nil {
            logger.Fatalf("Failed to initialize version strategy: %v", err)
        }
        
        e.Use(VersionLogger())
        e.Use(VersionValidator(versionConfig))
        e.Use(VersionDeprecation())
        
        logger.Infof("API versioning enabled - Strategy: %s, Default: %s", 
            versionConfig.Strategy, 
            versionConfig.DefaultVersion)
    }
}
```

### Step 3: Use in Routes (Works with Any Strategy)

The route registration code stays the same regardless of strategy:

```go
import "ichi-go/pkg/versioning"

func (c *AuthController) RegisterRoutesV1(serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
    // For semantic: creates /ichi-go/api/v1/auth
    // For date: creates /ichi-go/api/2026-01/auth
    // For date_daily: creates /ichi-go/api/20260115/auth
    vr := versioning.NewVersionedRoute(serviceName, "your-version", Domain)
    
    publicGroup := vr.Group(e)
    publicGroup.POST("/login", c.Login)
}
```

### Step 4: Use Helper Functions for Date-Based Versions

```go
// Date-based monthly
version := versioning.NewDateVersion(2026, 1) // "2026-01"
current := versioning.GetCurrentDateVersion() // Current month

// Date-based daily
version := versioning.NewDateDailyVersion(2026, 1, 15) // "20260115"
current := versioning.GetCurrentDateDailyVersion() // Today

// Use in route registration
vr := versioning.NewVersionedRoute(serviceName, current, Domain)
```

## Strategy Comparison

### When to Use Each Strategy

#### Semantic Versioning (`v1`, `v2`)
**Use When:**
- Clear breaking changes between versions
- Long-lived versions (6+ months)
- Feature-based releases
- Small number of versions (2-5)

**Pros:**
- Simple and clear
- Industry standard
- Easy for developers
- Good for major releases

**Cons:**
- No time information
- Manual version increments

**Example Companies:** Stripe, GitHub, Twitter API

---

#### Date-Based Monthly (`2026-01`)
**Use When:**
- Regular monthly releases
- Time-based deprecation
- Continuous deployment
- Many concurrent versions

**Pros:**
- Automatic chronological ordering
- Clear age of version
- Natural deprecation timeline
- Self-documenting

**Cons:**
- Possibly too many versions
- Less meaningful for breaking changes

**Example Companies:** Slack API, SendGrid

---

#### Date-Based Daily (`20260115`)
**Use When:**
- Daily/continuous deployments
- Rapid iteration
- Need precise version tracking
- Short version lifecycles

**Pros:**
- Very precise versioning
- Natural for CI/CD
- Easy automation

**Cons:**
- Many versions to manage
- Can be overwhelming

**Example Companies:** Twilio (uses similar approach)

---

#### Custom Strategies
**Use When:**
- Unique business requirements
- Internal conventions
- Marketing-driven versions
- Special naming needed

**Pros:**
- Complete flexibility
- Match business needs
- Can encode metadata

**Cons:**
- More complex
- Less familiar to developers
- Need custom tooling

## Complete Examples

### Example 1: E-commerce API with Monthly Releases

```yaml
# config.yaml
api:
  versioning:
    enabled: true
    strategy: "date"
    default_version: "2026-01"
    supported_versions:
      - "2025-11"  # November 2025 release
      - "2025-12"  # December 2025 release
      - "2026-01"  # January 2026 release (current)
    deprecation:
      header_enabled: true
      sunset_notification_days: 90
```

```go
// controller/product_routes.go
func (c *ProductController) RegisterRoutes(serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
    // Get current month version
    currentVersion := versioning.GetCurrentDateVersion()
    
    // Register current version
    c.registerVersion(serviceName, e, auth, currentVersion)
    
    // Also support last month for transition
    lastMonth := versioning.NewDateVersion(2025, 12)
    c.registerVersion(serviceName, e, auth, lastMonth)
}

func (c *ProductController) registerVersion(serviceName string, e *echo.Echo, auth *authenticator.Authenticator, version versioning.APIVersion) {
    vr := versioning.NewVersionedRoute(serviceName, version, "products")
    
    publicGroup := vr.Group(e)
    publicGroup.GET("", c.List)
    publicGroup.GET("/:id", c.GetByID)
    
    protectedGroup := vr.Group(e)
    protectedGroup.Use(auth.AuthenticateMiddleware())
    protectedGroup.POST("", c.Create)
}
```

```bash
# API Usage
curl http://localhost:8080/ichi-go/api/2026-01/products
curl http://localhost:8080/ichi-go/api/2025-12/products  # Still works
curl http://localhost:8080/ichi-go/api/2025-11/products  # Deprecated warning
```

### Example 2: SaaS Platform with Quarterly Releases

```yaml
# config.yaml
api:
  versioning:
    enabled: true
    strategy: "custom"
    default_version: "release-2026-Q1"
    custom_pattern: "^release-\\d{4}-Q[1-4]$"
    supported_versions:
      - "release-2025-Q3"
      - "release-2025-Q4"
      - "release-2026-Q1"
```

```go
// Define quarterly versions
const (
    Q3_2025 versioning.APIVersion = "release-2025-Q3"
    Q4_2025 versioning.APIVersion = "release-2025-Q4"
    Q1_2026 versioning.APIVersion = "release-2026-Q1"
)

func (c *AuthController) RegisterRoutes(serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
    // Register each quarter
    c.registerQuarterRoutes(serviceName, e, auth, Q1_2026) // Current
    c.registerQuarterRoutes(serviceName, e, auth, Q4_2025) // Previous
}
```

### Example 3: API with Named Versions

```yaml
# config.yaml
api:
  versioning:
    enabled: true
    strategy: "custom"
    default_version: "stable"
    custom_pattern: ""  # Empty pattern - use explicit list
    custom_valid_versions:
      - "alpha"
      - "beta"
      - "stable"
      - "legacy"
```

```bash
# Endpoints
POST /ichi-go/api/stable/auth/login    # Production users
POST /ichi-go/api/beta/auth/login      # Beta testers
POST /ichi-go/api/alpha/auth/login     # Early access
POST /ichi-go/api/legacy/auth/login    # Deprecated
```

## Migration Between Strategies

### From Semantic to Date-Based

1. **Add date strategy support:**
```yaml
api:
  versioning:
    enabled: true
    strategy: "date"
    default_version: "2026-01"
    supported_versions: ["2026-01"]
```

2. **Run both temporarily:**
```go
// Support both during transition
vr1 := versioning.NewVersionedRoute(serviceName, "v1", Domain)
vr2 := versioning.NewVersionedRoute(serviceName, "2026-01", Domain)

// Register routes for both versions
c.registerWithRoute(vr1, e, auth)  // Old: /api/v1/auth/login
c.registerWithRoute(vr2, e, auth)  // New: /api/2026-01/auth/login
```

3. **Deprecate semantic versions:**
```go
versioning.DeprecationSchedule[versioning.V1] = &versioning.DeprecationInfo{
    Version:            "v1",
    DeprecatedAt:       time.Now(),
    SunsetDate:         time.Now().Add(90 * 24 * time.Hour),
    ReplacementVersion: "2026-01",
    Message:            "Migrating to date-based versioning",
}
```

## Best Practices

### 1. Choose One Strategy and Stick With It
Don't mix strategies (e.g., don't have both `v1` and `2026-01` as active versions)

### 2. For Date-Based: Auto-Generate Versions
```go
// Generate supported versions for last 3 months
func generateSupportedVersions() []string {
    versions := []string{}
    now := time.Now()
    
    for i := 0; i < 3; i++ {
        month := now.AddDate(0, -i, 0)
        version := versioning.NewDateVersion(month.Year(), int(month.Month()))
        versions = append(versions, string(version))
    }
    
    return versions
}
```

### 3. Document Your Strategy Choice
Add to your API documentation:

```markdown
## API Versioning

This API uses date-based monthly versioning (YYYY-MM format).

Current version: `2026-01`
Base URL: `https://api.example.com/ichi-go/api/2026-01/`

Versions are released monthly and supported for 3 months.
```

### 4. Automate Version Updates
For date-based versioning, automatically update default version:

```go
// In CI/CD or deployment script
func updateConfigForNewMonth() {
    currentMonth := versioning.GetCurrentDateVersion()
    // Update config.yaml with new default_version
}
```

## Testing Different Strategies

```go
func TestWithDifferentStrategies(t *testing.T) {
    // Save original
    original := versioning.GetVersionStrategy()
    defer versioning.SetVersionStrategy(original)
    
    // Test with semantic
    versioning.SetVersionStrategy(&versioning.SemanticVersionStrategy{})
    version, _ := versioning.ParseVersion("v1")
    assert.Equal(t, "v1", version.String())
    
    // Test with date
    versioning.SetVersionStrategy(&versioning.DateVersionStrategy{})
    version, _ = versioning.ParseVersion("2026-01")
    assert.Equal(t, "2026-01", version.String())
}
```

## FAQ

**Q: Can I change strategy after deployment?**
A: Yes, but you'll need a transition period supporting both strategies.

**Q: Which strategy is best?**
A: Semantic for major releases, date-based for continuous deployment.

**Q: Can I have custom date formats?**
A: Yes, use custom strategy with your own pattern.

**Q: How do I deprecate date-based versions?**
A: Same as semantic - add to DeprecationSchedule with dates.

**Q: Can I mix strategies?**
A: Technically yes, but not recommended. Choose one.

## Summary

The versioning package now supports:
- ✅ Multiple built-in strategies
- ✅ Custom pattern support
- ✅ Runtime strategy switching
- ✅ All strategies work with existing route code
- ✅ Automatic validation per strategy
- ✅ Helper functions for date-based versions

Choose the strategy that fits your release cadence and stick with it!