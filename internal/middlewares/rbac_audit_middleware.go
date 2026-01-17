package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"ichi-go/internal/applications/rbac/models"
	"ichi-go/internal/applications/rbac/repositories"
	"ichi-go/pkg/logger"
	"ichi-go/pkg/requestctx"

	"github.com/labstack/echo/v4"
)

// RBACAuditMiddleware logs RBAC-related operations to the audit log
// This middleware captures all mutation operations (create, update, delete)
// on RBAC resources and logs them for compliance and security auditing
func RBACAuditMiddleware(auditRepo *repositories.AuditRepository, config AuditConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			rc := requestctx.FromContext(ctx)

			// Skip audit for excluded paths
			if isExcludedPath(c.Path(), config.ExcludedPaths) {
				return next(c)
			}

			// Skip read-only operations if configured
			if config.MutationsOnly && c.Request().Method == http.MethodGet {
				return next(c)
			}

			// Capture request start time
			startTime := time.Now()

			// Capture request body for mutations
			var requestBody map[string]interface{}
			if shouldCaptureBody(c.Request().Method) {
				requestBody = captureRequestBody(c)
			}

			// Create response writer wrapper to capture status code
			resWrapper := &responseWrapper{
				ResponseWriter: c.Response().Writer,
				statusCode:     http.StatusOK,
			}
			c.Response().Writer = resWrapper

			// Execute request
			err := next(c)

			// Calculate latency
			latency := time.Since(startTime)
			latencyMs := int(latency.Milliseconds())

			// Determine if this is an RBAC mutation operation
			if shouldAudit(c, config) {
				// Create audit log entry
				log := &models.AuditLog{
					EventID:      generateEventID(rc.RequestID, rc.UserID),
					Timestamp:    startTime,
					ActorID:      rc.UserID,
					ActorType:    models.ActorTypeUser,
					Action:       mapHTTPMethodToAction(c.Request().Method),
					ResourceType: stringPtr(extractResourceType(c.Path())),
					ResourceID:   stringPtr(extractResourceID(c)),
					TenantID:     rc.TenantID,
					IPAddress:    stringPtr(rc.ClientIP),
					UserAgent:    stringPtr(rc.UserAgent),
					RequestID:    stringPtr(rc.RequestID),
					LatencyMs:    &latencyMs,
				}

				// Add request body as "before" state for mutations
				if requestBody != nil {
					log.PolicyBefore = requestBody
				}

				// Add response status
				statusCode := resWrapper.statusCode
				if err != nil {
					if he, ok := err.(*echo.HTTPError); ok {
						statusCode = he.Code
					}
				}

				// Set decision based on HTTP status
				if statusCode >= 200 && statusCode < 300 {
					log.Decision = stringPtr(models.DecisionAllow)
					log.DecisionReason = stringPtr("Operation completed successfully")
				} else {
					log.Decision = stringPtr(models.DecisionDeny)
					if err != nil {
						log.DecisionReason = stringPtr(err.Error())
					} else {
						log.DecisionReason = stringPtr("Operation failed")
					}
				}

				// Save audit log asynchronously
				go func() {
					if err := auditRepo.Create(ctx, log); err != nil {
						logger.Errorf("Failed to save audit log: %v", err)
					}
				}()
			}

			return err
		}
	}
}

// AuditConfig configures audit logging behavior
type AuditConfig struct {
	// MutationsOnly logs only create/update/delete operations
	MutationsOnly bool

	// ExcludedPaths are paths that skip audit logging
	ExcludedPaths []string

	// IncludedPaths are specific paths to audit (if empty, all paths are audited)
	IncludedPaths []string

	// CaptureRequestBody includes request body in audit log
	CaptureRequestBody bool

	// CaptureResponseBody includes response body in audit log
	CaptureResponseBody bool

	// MaxBodySize limits the size of captured request/response bodies
	MaxBodySize int64
}

// DefaultAuditConfig returns default audit configuration
func DefaultAuditConfig() AuditConfig {
	return AuditConfig{
		MutationsOnly:       true,
		ExcludedPaths:       []string{"/health", "/metrics"},
		IncludedPaths:       []string{},
		CaptureRequestBody:  true,
		CaptureResponseBody: false,
		MaxBodySize:         10 * 1024, // 10KB
	}
}

// responseWrapper wraps http.ResponseWriter to capture status code
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (w *responseWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	if w.body != nil {
		w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// shouldCaptureBody determines if request body should be captured
func shouldCaptureBody(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch
}

// captureRequestBody reads and captures the request body
func captureRequestBody(c echo.Context) map[string]interface{} {
	if c.Request().Body == nil {
		return nil
	}

	// Read body
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return nil
	}

	// Restore body for downstream handlers
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Parse JSON
	var body map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		// Not JSON, store as string
		return map[string]interface{}{
			"raw": string(bodyBytes),
		}
	}

	return body
}

// shouldAudit determines if a request should be audited
func shouldAudit(c echo.Context, config AuditConfig) bool {
	path := c.Path()

	// Check included paths first (if specified)
	if len(config.IncludedPaths) > 0 {
		included := false
		for _, p := range config.IncludedPaths {
			if matchPath(path, p) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	// Check excluded paths
	for _, p := range config.ExcludedPaths {
		if matchPath(path, p) {
			return false
		}
	}

	// Only audit RBAC paths
	return isRBACPath(path)
}

// isRBACPath checks if a path is an RBAC-related path
func isRBACPath(path string) bool {
	rbacPrefixes := []string{
		"/api/v1/rbac/",
		"/api/v1/roles/",
		"/api/v1/permissions/",
	}

	for _, prefix := range rbacPrefixes {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// matchPath checks if a path matches a pattern (simple prefix matching)
func matchPath(path, pattern string) bool {
	if pattern == path {
		return true
	}
	// Support wildcard suffix
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(path) >= len(prefix) && path[:len(prefix)] == prefix
	}
	return false
}

// mapHTTPMethodToAction maps HTTP methods to RBAC actions
func mapHTTPMethodToAction(method string) string {
	switch method {
	case http.MethodGet:
		return "audit_view"
	case http.MethodPost:
		return models.ActionPolicyAdded
	case http.MethodPut, http.MethodPatch:
		return "policy_modified"
	case http.MethodDelete:
		return models.ActionPolicyRemoved
	default:
		return "unknown"
	}
}

// extractResourceType extracts the resource type from the URL path
func extractResourceType(path string) string {
	// Example: /api/v1/rbac/roles/123 -> roles
	segments := splitPath(path)

	for i, seg := range segments {
		if seg == "rbac" && i+1 < len(segments) {
			return segments[i+1]
		}
	}

	if len(segments) >= 4 {
		return segments[3] // /api/v1/rbac/{resource}
	}

	return "unknown"
}

// extractResourceID extracts the resource ID from the URL path or params
func extractResourceID(c echo.Context) string {
	// Try common ID parameter names
	params := []string{"id", "roleId", "userId", "permissionId"}

	for _, param := range params {
		if id := c.Param(param); id != "" {
			return id
		}
	}

	return ""
}

// generateEventID generates a unique event ID
func generateEventID(requestID, userID string) string {
	if requestID != "" {
		return requestID
	}
	return userID + "_" + time.Now().Format("20060102150405")
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
