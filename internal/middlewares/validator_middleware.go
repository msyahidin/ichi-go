package middlewares

import (
	"fmt"
	"ichi-go/internal/applications/auth/validators"
	"ichi-go/pkg/logger"
	appValidator "ichi-go/pkg/validator"

	"github.com/labstack/echo/v4"
)

// ValidatorMiddleware provides Echo with our custom validator
type ValidatorMiddleware struct {
	validator *appValidator.AppValidator
}

// NewValidatorMiddleware creates a new validator middleware
func NewValidatorMiddleware(validator *appValidator.AppValidator) *ValidatorMiddleware {
	return &ValidatorMiddleware{
		validator: validator,
	}
}

// Validate implements echo.Validator interface
func (vm *ValidatorMiddleware) Validate(i interface{}) error {
	// Note: Echo passes the bound struct, not the context
	// We'll use default language here, language detection happens in controller
	validationErr := vm.validator.ValidateStruct(i)
	if validationErr != nil && validationErr.HasErrors() {
		return validationErr
	}
	return nil
}

// ValidateWithLanguage validates with specific language
func (vm *ValidatorMiddleware) ValidateWithLanguage(i interface{}, lang string) error {
	validationErr := vm.validator.Validate(i, lang)
	if validationErr != nil && validationErr.HasErrors() {
		return validationErr
	}
	return nil
}

// ValidateWithContext validates using language from Echo context
func (vm *ValidatorMiddleware) ValidateWithContext(c echo.Context, i interface{}) error {
	lang := appValidator.GetLanguageFromContext(c)
	return vm.ValidateWithLanguage(i, lang)
}

// setupValidator creates and configures the validator with all domain validators
func setupValidator(e *echo.Echo, config appValidator.Config) error {
	// Create base validator with translation support
	v, err := appValidator.NewValidator(config)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	// Register auth domain validators
	if err := validators.RegisterAuthValidators(v); err != nil {
		return fmt.Errorf("failed to register auth validators: %w", err)
	}

	// Register other domain validators here:
	// if err := userValidators.RegisterUserValidators(v); err != nil {
	//     return fmt.Errorf("failed to register user validators: %w", err)
	// }

	// Set Echo validator
	e.Validator = NewValidatorMiddleware(v)

	logger.Debugf("Validator initialized with auth validators")
	return nil
}
