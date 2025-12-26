package middlewares

import (
	"fmt"
	"github.com/labstack/echo/v4"
	authValidators "ichi-go/internal/applications/auth/validators"
	appValidator "ichi-go/pkg/validator"
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

// InitValidator sets up the validator with all domain validators and registers it with Echo
func InitValidator(e *echo.Echo, config appValidator.Config) error {
	// Create base validator with translation support
	v, err := appValidator.NewValidator(config)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	// Register auth domain validators (strong_password, username_format)
	if err := authValidators.RegisterAuthValidators(v); err != nil {
		return fmt.Errorf("failed to register auth validators: %w", err)
	}

	// Add more domain validators here as your application grows:
	//
	// if err := productValidators.RegisterProductValidators(v); err != nil {
	//     return fmt.Errorf("failed to register product validators: %w", err)
	// }

	// Set Echo validator
	e.Validator = NewValidatorMiddleware(v)

	return nil
}
