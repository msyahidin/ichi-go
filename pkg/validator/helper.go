package validator

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrValidationFailed is a sentinel error for validation failures
var ErrValidationFailed = errors.New("validation failed")

type ContextValidator interface {
	ValidateWithContext(c echo.Context, i interface{}) error
	ValidateWithLanguage(i interface{}, lang string) error
}

// BindAndValidate binds request body and validates with automatic language detection
// Returns ValidationErrors that can be passed to response.Error()
func BindAndValidate(c echo.Context, req interface{}) error {
	// Bind request body
	if err := c.Bind(req); err != nil {
		return err
	}

	// Validate with language detection
	return ValidateWithTranslation(c, req)
}

// ValidateWithTranslation validates a struct with automatic language detection from headers
// Returns ValidationErrors without sending response (compatible with response builder)
func ValidateWithTranslation(c echo.Context, req interface{}) error {
	// Get validator from Echo
	echoValidator := c.Echo().Validator
	if echoValidator == nil {
		return errors.New("validator not configured")
	}

	// Try to use context-aware validator
	if contextValidator, ok := echoValidator.(ContextValidator); ok {
		if err := contextValidator.ValidateWithContext(c, req); err != nil {
			// Return the ValidationErrors directly - don't send response
			return err
		}
		return nil
	}

	// Fallback to basic validation if custom validator not available
	if err := c.Validate(req); err != nil {
		return err
	}

	return nil
}

// ValidateWithLanguage validates a struct with a specific language
// Returns ValidationErrors without sending response (compatible with response builder)
func ValidateWithLanguage(c echo.Context, req interface{}, lang string) error {
	// Get validator from Echo
	echoValidator := c.Echo().Validator
	if echoValidator == nil {
		return errors.New("validator not configured")
	}

	// Try to use context-aware validator
	if contextValidator, ok := echoValidator.(ContextValidator); ok {
		if err := contextValidator.ValidateWithLanguage(req, lang); err != nil {
			return err
		}
		return nil
	}

	// Fallback to basic validation
	if err := c.Validate(req); err != nil {
		return err
	}

	return nil
}

// GetValidationError extracts ValidationErrors from error
// Returns nil if error is not a ValidationErrors type
func GetValidationError(err error) *ValidationErrors {
	if err == nil {
		return nil
	}

	validationErr, ok := err.(*ValidationErrors)
	if !ok {
		return nil
	}

	return validationErr
}

// FormatValidationResponse formats ValidationErrors as HTTP response map
func FormatValidationResponse(errors *ValidationErrors) map[string]interface{} {
	if errors == nil {
		return map[string]interface{}{
			"code":    http.StatusBadRequest,
			"message": "Validation failed",
		}
	}

	return map[string]interface{}{
		"code":    http.StatusBadRequest,
		"message": "Validation failed",
		"errors":  errors.Errors,
	}
}
