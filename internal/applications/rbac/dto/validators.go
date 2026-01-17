package dto

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// Custom validation tags for RBAC DTOs

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ValidateSlug validates that a string is a valid slug format
func ValidateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	return slugRegex.MatchString(slug)
}

// RegisterValidators registers custom validators for RBAC DTOs
func RegisterValidators(v *validator.Validate) error {
	if err := v.RegisterValidation("slug", ValidateSlug); err != nil {
		return err
	}
	return nil
}
