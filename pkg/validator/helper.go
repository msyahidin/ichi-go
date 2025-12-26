package validator

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type ContextValidator interface {
	ValidateWithContext(c echo.Context, i interface{}) error
	ValidateWithLanguage(i interface{}, lang string) error
}

// BindAndValidate binds request body and validates with automatic language detection
// Returns formatted JSON error response if validation fails, nil if successful
func BindAndValidate(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    http.StatusBadRequest,
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// Validate with language detection
	return ValidateWithTranslation(c, req)
}

// ValidateWithTranslation validates a struct with automatic language detection from headers
// Returns formatted JSON error response if validation fails, nil if successful
func ValidateWithTranslation(c echo.Context, req interface{}) error {
	// Get validator from Echo
	echoValidator := c.Echo().Validator
	if echoValidator == nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code":    http.StatusInternalServerError,
			"message": "Validator not configured",
		})
	}

	// Try to use context-aware validator
	if contextValidator, ok := echoValidator.(ContextValidator); ok {
		if err := contextValidator.ValidateWithContext(c, req); err != nil {
			// Check if it's our custom ValidationErrors type
			if validationErr, ok := err.(*ValidationErrors); ok {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"code":    http.StatusBadRequest,
					"message": "Validation failed",
					"errors":  validationErr.Errors,
				})
			}

			// Fallback for other error types
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": "Validation failed",
				"error":   err.Error(),
			})
		}
		return nil
	}

	// Fallback to basic validation if custom validator not available
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    http.StatusBadRequest,
			"message": "Validation failed",
			"error":   err.Error(),
		})
	}

	return nil
}

// ValidateWithLanguage validates a struct with a specific language
// Returns formatted JSON error response if validation fails, nil if successful
func ValidateWithLanguage(c echo.Context, req interface{}, lang string) error {
	// Get validator from Echo
	echoValidator := c.Echo().Validator
	if echoValidator == nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code":    http.StatusInternalServerError,
			"message": "Validator not configured",
		})
	}

	// Try to use context-aware validator
	if contextValidator, ok := echoValidator.(ContextValidator); ok {
		if err := contextValidator.ValidateWithLanguage(req, lang); err != nil {
			if validationErr, ok := err.(*ValidationErrors); ok {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"code":    http.StatusBadRequest,
					"message": "Validation failed",
					"errors":  validationErr.Errors,
				})
			}

			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": "Validation failed",
				"error":   err.Error(),
			})
		}
		return nil
	}

	// Fallback to basic validation
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    http.StatusBadRequest,
			"message": "Validation failed",
			"error":   err.Error(),
		})
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
