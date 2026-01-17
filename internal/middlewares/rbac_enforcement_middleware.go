package middlewares

import (
	"net/http"

	"ichi-go/internal/applications/rbac/services"
	"ichi-go/pkg/logger"
	"ichi-go/pkg/requestctx"

	"github.com/labstack/echo/v4"
)

// RBACEnforcementMiddleware enforces RBAC permissions on routes
// This middleware checks if the authenticated user has the required permissions
// to access the requested resource/action combination
func RBACEnforcementMiddleware(enforcementService *services.EnforcementService, config RBACConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip enforcement for excluded paths
			if isExcludedPath(c.Path(), config.ExcludedPaths) {
				return next(c)
			}

			// Get request context
			ctx := c.Request().Context()
			rc := requestctx.FromContext(ctx)

			// Skip if guest and guests are allowed
			if rc.IsGuest && config.AllowGuests {
				return next(c)
			}

			// Get user ID
			userID := requestctx.GetUserIDAsInt64(ctx)
			if userID == 0 {
				logger.WithContext(ctx).Warn("RBAC enforcement: No user ID in context")
				return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
			}

			// Get tenant ID
			tenantID := requestctx.GetTenantID(ctx)
			if tenantID == "" && config.RequireTenant {
				logger.WithContext(ctx).Warn("RBAC enforcement: No tenant ID in context")
				return echo.NewHTTPError(http.StatusBadRequest, "Tenant context required")
			}

			// Use default tenant if not provided
			if tenantID == "" {
				tenantID = config.DefaultTenant
			}

			// Resolve resource and action from route
			resource, action := resolveResourceAction(c, config)

			// Check permission
			allowed, err := enforcementService.CheckPermission(
				ctx,
				userID,
				tenantID,
				resource,
				action,
			)

			if err != nil {
				logger.WithContext(ctx).Errorf("RBAC enforcement error: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Permission check failed")
			}

			if !allowed {
				logger.WithContext(ctx).Warnf(
					"RBAC permission denied: user=%d tenant=%s resource=%s action=%s",
					userID, tenantID, resource, action,
				)
				return echo.NewHTTPError(http.StatusForbidden, "Permission denied")
			}

			// Permission granted, continue
			logger.WithContext(ctx).Debugf(
				"RBAC permission granted: user=%d tenant=%s resource=%s action=%s",
				userID, tenantID, resource, action,
			)

			return next(c)
		}
	}
}

// RBACConfig configures RBAC enforcement behavior
type RBACConfig struct {
	// RequireTenant enforces that tenant context must be present
	RequireTenant bool

	// DefaultTenant is used when no tenant context is found
	DefaultTenant string

	// AllowGuests allows unauthenticated requests to pass through
	AllowGuests bool

	// ExcludedPaths are paths that skip RBAC enforcement
	ExcludedPaths []string

	// ResourceMapper is a custom function to map routes to resources
	ResourceMapper func(c echo.Context) (resource string, action string)

	// DefaultResource is used when ResourceMapper returns empty string
	DefaultResource string

	// DefaultAction is used when ResourceMapper returns empty action
	DefaultAction string
}

// DefaultRBACConfig returns default RBAC configuration
func DefaultRBACConfig() RBACConfig {
	return RBACConfig{
		RequireTenant:   false,
		DefaultTenant:   "",
		AllowGuests:     false,
		ExcludedPaths:   []string{"/health", "/api/v1/auth/login", "/api/v1/auth/register"},
		ResourceMapper:  defaultResourceMapper,
		DefaultResource: "api",
		DefaultAction:   "access",
	}
}

// RequirePermission creates a middleware that enforces a specific permission
// This is useful for protecting individual routes with explicit permissions
func RequirePermission(enforcementService *services.EnforcementService, resource, action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			// Get user ID
			userID := requestctx.GetUserIDAsInt64(ctx)
			if userID == 0 {
				return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
			}

			// Get tenant ID
			tenantID := requestctx.GetTenantID(ctx)
			if tenantID == "" {
				return echo.NewHTTPError(http.StatusBadRequest, "Tenant context required")
			}

			// Check permission
			allowed, err := enforcementService.CheckPermission(ctx, userID, tenantID, resource, action)
			if err != nil {
				logger.WithContext(ctx).Errorf("Permission check error: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Permission check failed")
			}

			if !allowed {
				logger.WithContext(ctx).Warnf(
					"Permission denied: user=%d tenant=%s resource=%s action=%s",
					userID, tenantID, resource, action,
				)
				return echo.NewHTTPError(http.StatusForbidden, "Permission denied")
			}

			return next(c)
		}
	}
}

// resolveResourceAction determines the resource and action from the request
func resolveResourceAction(c echo.Context, config RBACConfig) (resource, action string) {
	// Use custom mapper if provided
	if config.ResourceMapper != nil {
		resource, action = config.ResourceMapper(c)
	}

	// Use defaults if mapper returned empty values
	if resource == "" {
		resource = config.DefaultResource
	}
	if action == "" {
		action = config.DefaultAction
	}

	return resource, action
}

// defaultResourceMapper is the default resource/action mapper
// Maps HTTP methods to RBAC actions and extracts resource from path
func defaultResourceMapper(c echo.Context) (resource, action string) {
	// Map HTTP method to RBAC action
	switch c.Request().Method {
	case http.MethodGet:
		action = "view"
	case http.MethodPost:
		action = "create"
	case http.MethodPut, http.MethodPatch:
		action = "edit"
	case http.MethodDelete:
		action = "delete"
	default:
		action = "access"
	}

	// Extract resource from path
	// Example: /api/v1/users -> users
	path := c.Path()

	// Remove /api/v{version}/ prefix if present
	if len(path) > 8 && path[:7] == "/api/v" {
		// Find the next slash after version
		idx := 0
		for i := 8; i < len(path); i++ {
			if path[i] == '/' {
				idx = i + 1
				break
			}
		}
		if idx > 0 && idx < len(path) {
			path = path[idx:]
		}
	}

	// Get first path segment as resource
	segments := splitPath(path)
	if len(segments) > 0 {
		resource = segments[0]
	}

	return resource, action
}

// splitPath splits a path into segments
func splitPath(path string) []string {
	var segments []string
	start := 0

	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				segments = append(segments, path[start:i])
			}
			start = i + 1
		}
	}

	// Add last segment
	if start < len(path) {
		segments = append(segments, path[start:])
	}

	return segments
}

// isExcludedPath checks if a path is in the excluded list
func isExcludedPath(path string, excluded []string) bool {
	for _, ex := range excluded {
		if path == ex {
			return true
		}
	}
	return false
}
