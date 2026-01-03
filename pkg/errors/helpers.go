package errors

import (
	"github.com/samber/oops"
)

// ============================================================================
// Generic Builders
// ============================================================================

// New creates a new oops error
func New(in string, code string) oops.OopsErrorBuilder {
	return oops.Code(code).In(in)
}

// Wrap wraps an existing error with context
func Wrap(in string, code string, err error, context map[string]interface{}) error {
	builder := oops.Code(code).In(in)

	for k, v := range context {
		builder = builder.With(k, v)
	}

	return builder.Wrap(err)
}

// ============================================================================
// Service Layer Errors
// ============================================================================

// AuthService - Authentication/Authorization operations
func AuthService(code string) oops.OopsErrorBuilder {
	return oops.Code(code).
		In("auth-service").
		Tags("authentication")
}

// UserService - User management operations
func UserService(code string) oops.OopsErrorBuilder {
	return oops.Code(code).
		In("user-service").
		Tags("user-management")
}

// OrderService - Order processing operations
func OrderService(code string) oops.OopsErrorBuilder {
	return oops.Code(code).
		In("order-service").
		Tags("commerce")
}

// PaymentService - Payment processing operations
func PaymentService(code string) oops.OopsErrorBuilder {
	return oops.Code(code).
		In("payment-service").
		Tags("payments")
}

// ============================================================================
// Infrastructure Layer Errors
// ============================================================================

// Database - Database operations
func Database(code string) oops.OopsErrorBuilder {
	return oops.Code(code).
		In("database").
		Tags("infrastructure", "persistence")
}

// Cache - Redis/cache operations
func Cache(code string) oops.OopsErrorBuilder {
	return oops.Code(code).
		In("cache").
		Tags("infrastructure", "caching")
}

// Queue - Message queue operations
func Queue(code string) oops.OopsErrorBuilder {
	return oops.Code(code).
		In("queue").
		Tags("infrastructure", "messaging")
}

// ExternalAPI - External API calls
func ExternalAPI(code string) oops.OopsErrorBuilder {
	return oops.Code(code).
		In("external-api").
		Tags("integration")
}

// ============================================================================
// Generic Helpers
// ============================================================================

// Internal - Generic internal error
func Internal(domain string, code string) oops.OopsErrorBuilder {
	return oops.Code(code).
		In(domain).
		Tags("internal")
}

// Validation - Validation errors
func Validation(code string) oops.OopsErrorBuilder {
	return oops.Code(code).
		In("validator").
		Tags("validation")
}
