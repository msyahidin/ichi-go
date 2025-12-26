package validator

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// FieldError represents a single validation error
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors holds all validation errors
type ValidationErrors struct {
	Errors []FieldError `json:"errors"`
}

// Error implements error interface
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "validation failed"
	}
	return ve.Errors[0].Message
}

// HasErrors checks if there are any errors
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// GetFirstError returns the first error message
func (ve *ValidationErrors) GetFirstError() string {
	if len(ve.Errors) == 0 {
		return ""
	}
	return ve.Errors[0].Message
}

// GetFieldErrors returns all field errors as a map
func (ve *ValidationErrors) GetFieldErrors() map[string]string {
	fieldErrors := make(map[string]string)
	for _, err := range ve.Errors {
		fieldErrors[err.Field] = err.Message
	}
	return fieldErrors
}

// FormatValidationErrors converts validator errors to our format
func FormatValidationErrors(errs validator.ValidationErrors, trans ut.Translator) *ValidationErrors {
	validationErrors := &ValidationErrors{
		Errors: make([]FieldError, 0, len(errs)),
	}

	for _, err := range errs {
		validationErrors.Errors = append(validationErrors.Errors, FieldError{
			Field:   err.Field(),
			Message: err.Translate(trans),
			Tag:     err.Tag(),
			Value:   err.Param(),
		})
	}

	return validationErrors
}

// ToMap converts ValidationErrors to a simple map for JSON responses
func (ve *ValidationErrors) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"validation_errors": ve.Errors,
	}
}
