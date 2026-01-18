package middlewares

import (
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
