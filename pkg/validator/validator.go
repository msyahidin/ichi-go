package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator wraps the validator instance
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates the struct
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

// NewValidator creates a new validator instance
func NewValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

// RegisterValidator registers the validator with Echo
func RegisterValidator(e *echo.Echo) {
	e.Validator = NewValidator()
}
