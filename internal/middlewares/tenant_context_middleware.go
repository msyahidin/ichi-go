package middlewares

import (
	"net/http"
	"strings"

	"ichi-go/pkg/logger"
	"ichi-go/pkg/requestctx"

	"github.com/labstack/echo/v4"
)

// TenantContextMiddleware extracts and validates tenant context from requests
// Supports multiple tenant resolution strategies:
// 1. X-Tenant-Id header (explicit)
// 2. Subdomain extraction (tenant.example.com -> tenant)
// 3. Path prefix (/tenants/{tenantId}/...)
// 4. Query parameter (?tenant_id=...)
func TenantContextMiddleware(config TenantConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get request context
			rc := requestctx.FromContext(c.Request().Context())
			if rc == nil {
				rc = requestctx.FromRequest(c.Request())
			}

			// Resolve tenant ID using configured strategy
			tenantID := resolveTenantID(c, config)

			// Set tenant ID in request context
			rc.TenantID = tenantID

			// Update request context
			ctx := requestctx.NewContext(c.Request().Context(), rc)
			c.SetRequest(c.Request().WithContext(ctx))

			// Log tenant context (debug)
			if tenantID != "" {
				logger.WithContext(ctx).Debugf("Tenant context: %s", tenantID)
			} else if config.RequireTenant {
				logger.WithContext(ctx).Warn("No tenant context found in request")
			}

			// Enforce tenant requirement if configured
			if config.RequireTenant && tenantID == "" {
				return echo.NewHTTPError(http.StatusBadRequest, "Tenant context is required")
			}

			return next(c)
		}
	}
}

// TenantConfig configures tenant resolution behavior
type TenantConfig struct {
	// RequireTenant enforces that all requests must have a tenant context
	RequireTenant bool

	// Strategy determines how tenant ID is resolved
	// Options: "header", "subdomain", "path", "query", "auto"
	Strategy string

	// HeaderName is the header key for tenant ID (default: X-Tenant-Id)
	HeaderName string

	// SubdomainPosition is the position of tenant in subdomain (0 = first)
	SubdomainPosition int

	// PathPrefix is the URL path prefix for tenant extraction
	// Example: "/tenants/{tenantId}" -> PathPrefix = "/tenants/"
	PathPrefix string

	// QueryParam is the query parameter name for tenant ID
	QueryParam string

	// DefaultTenant is used when no tenant is found and RequireTenant is false
	DefaultTenant string

	// AllowedTenants is a whitelist of valid tenant IDs (optional)
	AllowedTenants []string
}

// DefaultTenantConfig returns default configuration
func DefaultTenantConfig() TenantConfig {
	return TenantConfig{
		RequireTenant:     false,
		Strategy:          "auto", // Try all strategies
		HeaderName:        "X-Tenant-Id",
		SubdomainPosition: 0,
		PathPrefix:        "/tenants/",
		QueryParam:        "tenant_id",
		DefaultTenant:     "",
		AllowedTenants:    nil,
	}
}

// resolveTenantID resolves tenant ID based on configured strategy
func resolveTenantID(c echo.Context, config TenantConfig) string {
	var tenantID string

	switch config.Strategy {
	case "header":
		tenantID = resolveTenantFromHeader(c, config)
	case "subdomain":
		tenantID = resolveTenantFromSubdomain(c, config)
	case "path":
		tenantID = resolveTenantFromPath(c, config)
	case "query":
		tenantID = resolveTenantFromQuery(c, config)
	case "auto":
		// Try strategies in order: header -> subdomain -> path -> query
		tenantID = resolveTenantFromHeader(c, config)
		if tenantID == "" {
			tenantID = resolveTenantFromSubdomain(c, config)
		}
		if tenantID == "" {
			tenantID = resolveTenantFromPath(c, config)
		}
		if tenantID == "" {
			tenantID = resolveTenantFromQuery(c, config)
		}
	default:
		tenantID = resolveTenantFromHeader(c, config)
	}

	// Validate against whitelist if configured
	if len(config.AllowedTenants) > 0 && tenantID != "" {
		if !isAllowedTenant(tenantID, config.AllowedTenants) {
			logger.Warnf("Tenant '%s' not in allowed list", tenantID)
			return ""
		}
	}

	// Use default tenant if no tenant found
	if tenantID == "" && config.DefaultTenant != "" {
		tenantID = config.DefaultTenant
	}

	return tenantID
}

// resolveTenantFromHeader extracts tenant from HTTP header
func resolveTenantFromHeader(c echo.Context, config TenantConfig) string {
	headerName := config.HeaderName
	if headerName == "" {
		headerName = "X-Tenant-Id"
	}
	return c.Request().Header.Get(headerName)
}

// resolveTenantFromSubdomain extracts tenant from subdomain
// Example: tenant1.example.com -> tenant1
func resolveTenantFromSubdomain(c echo.Context, config TenantConfig) string {
	host := c.Request().Host

	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Split by dots
	parts := strings.Split(host, ".")

	// Need at least 3 parts for subdomain (subdomain.domain.tld)
	if len(parts) < 3 {
		return ""
	}

	// Get subdomain at configured position
	if config.SubdomainPosition < 0 || config.SubdomainPosition >= len(parts)-2 {
		return ""
	}

	return parts[config.SubdomainPosition]
}

// resolveTenantFromPath extracts tenant from URL path
// Example: /tenants/acme/users -> acme
func resolveTenantFromPath(c echo.Context, config TenantConfig) string {
	path := c.Request().URL.Path

	if !strings.HasPrefix(path, config.PathPrefix) {
		return ""
	}

	// Remove prefix
	remaining := strings.TrimPrefix(path, config.PathPrefix)

	// Get first segment
	segments := strings.Split(remaining, "/")
	if len(segments) > 0 && segments[0] != "" {
		return segments[0]
	}

	return ""
}

// resolveTenantFromQuery extracts tenant from query parameter
func resolveTenantFromQuery(c echo.Context, config TenantConfig) string {
	paramName := config.QueryParam
	if paramName == "" {
		paramName = "tenant_id"
	}
	return c.QueryParam(paramName)
}

// isAllowedTenant checks if tenant is in whitelist
func isAllowedTenant(tenantID string, allowed []string) bool {
	for _, t := range allowed {
		if t == tenantID {
			return true
		}
	}
	return false
}
