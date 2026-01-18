package constants

import "errors"

// Domain errors for RBAC operations

var (
	// Role errors
	ErrRoleNotFound      = errors.New("role not found")
	ErrRoleAlreadyExists = errors.New("role already exists")
	ErrRoleIsSystemRole  = errors.New("cannot modify system role")
	ErrRoleHasUsers      = errors.New("role has assigned users")
	ErrInvalidRoleSlug   = errors.New("invalid role slug format")

	// Permission errors
	ErrPermissionNotFound      = errors.New("permission not found")
	ErrPermissionAlreadyExists = errors.New("permission already exists")
	ErrInvalidPermissionSlug   = errors.New("invalid permission slug format")

	// User role errors
	ErrUserRoleNotFound      = errors.New("user role assignment not found")
	ErrUserRoleAlreadyExists = errors.New("user already has this role")
	ErrUserRoleExpired       = errors.New("user role assignment has expired")

	// Policy errors
	ErrPolicyNotFound      = errors.New("policy not found")
	ErrPolicyAlreadyExists = errors.New("policy already exists")
	ErrInvalidPolicy       = errors.New("invalid policy format")
	ErrPolicyLoadFailed    = errors.New("failed to load policies")

	// Audit errors
	ErrAuditLogNotFound   = errors.New("audit log not found")
	ErrAuditExportFailed  = errors.New("failed to export audit logs")
	ErrAuditCleanupFailed = errors.New("failed to cleanup old audit logs")

	// Enforcement errors
	ErrEnforcementFailed = errors.New("permission check failed")
	ErrCacheMiss         = errors.New("cache miss")
	ErrCacheInvalidation = errors.New("failed to invalidate cache")

	// Tenant errors
	ErrTenantNotFound           = errors.New("tenant not found")
	ErrInvalidTenantID          = errors.New("invalid tenant ID")
	ErrTenantIsolationViolation = errors.New("tenant isolation violation")

	// Validation errors
	ErrValidationFailed     = errors.New("validation failed")
	ErrMissingRequiredField = errors.New("missing required field")
	ErrInvalidFieldValue    = errors.New("invalid field value")

	// Authorization errors
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrPermissionDenied = errors.New("permission denied")

	// Database errors
	ErrDatabaseConnection  = errors.New("database connection failed")
	ErrDatabaseQuery       = errors.New("database query failed")
	ErrDatabaseTransaction = errors.New("database transaction failed")

	// Circuit breaker errors
	ErrCircuitBreakerOpen    = errors.New("circuit breaker is open")
	ErrCircuitBreakerTimeout = errors.New("circuit breaker timeout")
)

// Error codes for API responses
const (
	// Role error codes
	CodeRoleNotFound      = "ROLE_NOT_FOUND"
	CodeRoleAlreadyExists = "ROLE_ALREADY_EXISTS"
	CodeRoleIsSystemRole  = "ROLE_IS_SYSTEM_ROLE"
	CodeRoleHasUsers      = "ROLE_HAS_USERS"

	// Permission error codes
	CodePermissionNotFound      = "PERMISSION_NOT_FOUND"
	CodePermissionAlreadyExists = "PERMISSION_ALREADY_EXISTS"

	// User role error codes
	CodeUserRoleNotFound      = "USER_ROLE_NOT_FOUND"
	CodeUserRoleAlreadyExists = "USER_ROLE_ALREADY_EXISTS"
	CodeUserRoleExpired       = "USER_ROLE_EXPIRED"

	// Policy error codes
	CodePolicyNotFound      = "POLICY_NOT_FOUND"
	CodePolicyAlreadyExists = "POLICY_ALREADY_EXISTS"
	CodeInvalidPolicy       = "INVALID_POLICY"

	// Enforcement error codes
	CodeEnforcementFailed = "ENFORCEMENT_FAILED"
	CodePermissionDenied  = "PERMISSION_DENIED"

	// Tenant error codes
	CodeTenantNotFound           = "TENANT_NOT_FOUND"
	CodeInvalidTenantID          = "INVALID_TENANT_ID"
	CodeTenantIsolationViolation = "TENANT_ISOLATION_VIOLATION"

	// Validation error codes
	CodeValidationFailed     = "VALIDATION_FAILED"
	CodeMissingRequiredField = "MISSING_REQUIRED_FIELD"
	CodeInvalidFieldValue    = "INVALID_FIELD_VALUE"

	// Authorization error codes
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"

	// System error codes
	CodeInternalError = "INTERNAL_ERROR"
	CodeDatabaseError = "DATABASE_ERROR"
)
