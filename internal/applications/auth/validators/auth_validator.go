package validators

import (
	"regexp"
	"unicode"

	appValidator "ichi-go/pkg/validator"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// RegisterAuthValidators registers authentication-specific validators
func RegisterAuthValidators(av *appValidator.AppValidator) error {
	// Get the underlying validator
	v := av.GetValidator()

	// Register validation functions
	if err := v.RegisterValidation("strong_password", validateStrongPassword); err != nil {
		return err
	}

	if err := v.RegisterValidation("username_format", validateUsernameFormat); err != nil {
		return err
	}

	// Register translations for each supported language
	for _, lang := range []string{"en", "id"} {
		trans := av.GetTranslator(lang)

		if err := registerPasswordTranslation(v, trans, lang); err != nil {
			return err
		}

		if err := registerUsernameTranslation(v, trans, lang); err != nil {
			return err
		}
	}

	return nil
}

// validateStrongPassword validates password strength
// Requirements:
// - At least 8 characters
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one number
// - At least one special character
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
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

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// validateUsernameFormat validates username format
// Requirements:
// - 3-20 characters
// - Only alphanumeric, underscore, hyphen
// - Must start with letter
// - Cannot end with underscore or hyphen
func validateUsernameFormat(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	if len(username) < 3 || len(username) > 20 {
		return false
	}

	// Regex: starts with letter, contains alphanumeric/underscore/hyphen, doesn't end with underscore/hyphen
	pattern := `^[a-zA-Z][a-zA-Z0-9_-]*[a-zA-Z0-9]$`
	matched, _ := regexp.MatchString(pattern, username)
	return matched
}

// registerPasswordTranslation registers password validator translation
func registerPasswordTranslation(v *validator.Validate, trans ut.Translator, lang string) error {
	var message string

	switch lang {
	case "id":
		message = "{0} harus berisi minimal 8 karakter, termasuk huruf besar, huruf kecil, angka, dan karakter khusus"
	default: // "en"
		message = "{0} must contain at least 8 characters, including uppercase, lowercase, number, and special character"
	}

	return v.RegisterTranslation("strong_password", trans,
		func(ut ut.Translator) error {
			return ut.Add("strong_password", message, true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("strong_password", fe.Field())
			return t
		},
	)
}

// registerUsernameTranslation registers username validator translation
func registerUsernameTranslation(v *validator.Validate, trans ut.Translator, lang string) error {
	var message string

	switch lang {
	case "id":
		message = "{0} harus berisi 3-20 karakter, dimulai dengan huruf, dan hanya boleh mengandung huruf, angka, garis bawah, atau tanda hubung"
	default: // "en"
		message = "{0} must be 3-20 characters, start with a letter, and contain only letters, numbers, underscores, or hyphens"
	}

	return v.RegisterTranslation("username_format", trans,
		func(ut ut.Translator) error {
			return ut.Add("username_format", message, true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("username_format", fe.Field())
			return t
		},
	)
}
