package validators

import (
	"regexp"
	"unicode"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	appValidator "ichi-go/pkg/validator"
)

// RegisterAuthValidators registers all auth-related custom validators
func RegisterAuthValidators(av *appValidator.AppValidator) error {
	validators := []appValidator.CustomValidator{
		{
			Tag:           "strong_password",
			Fn:            validateStrongPassword,
			RegisterTrans: registerPasswordTranslations,
		},
		{
			Tag:           "username_format",
			Fn:            validateUsernameFormat,
			RegisterTrans: registerUsernameTranslations,
		},
	}

	return av.RegisterCustomValidators(validators)
}

// validateStrongPassword checks password strength
// Requirements:
// - At least 8 characters
// - At least 1 uppercase letter
// - At least 1 lowercase letter
// - At least 1 number
// - At least 1 special character
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	var (
		hasMinLen  = len(password) >= 8
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

// validateUsernameFormat validates username format
// Requirements:
// - Must start with a letter
// - Only alphanumeric, underscore, and hyphen allowed
// - 3-50 characters
func validateUsernameFormat(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// Check length
	if len(username) < 3 || len(username) > 50 {
		return false
	}

	// Only alphanumeric, underscore, and hyphen
	// Must start with letter
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_-]*$`, username)
	return matched
}

// registerPasswordTranslations registers translations for password validation
func registerPasswordTranslations(trans ut.Translator) error {
	messages := map[string]string{
		"en": "{0} must contain at least 8 characters, including uppercase, lowercase, number, and special character",
		"id": "{0} harus berisi minimal 8 karakter, termasuk huruf besar, huruf kecil, angka, dan karakter khusus",
	}

	locale := trans.Locale()
	message, ok := messages[locale]
	if !ok {
		message = messages["en"]
	}

	return trans.Add("strong_password", message, true)
}

// registerUsernameTranslations registers translations for username validation
func registerUsernameTranslations(trans ut.Translator) error {
	messages := map[string]string{
		"en": "{0} must start with a letter and contain only letters, numbers, underscore, and hyphen (3-50 characters)",
		"id": "{0} harus diawali dengan huruf dan hanya berisi huruf, angka, underscore, dan tanda hubung (3-50 karakter)",
	}

	locale := trans.Locale()
	message, ok := messages[locale]
	if !ok {
		message = messages["en"]
	}

	return trans.Add("username_format", message, true)
}
