package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	config := Config{
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en", "id"},
	}

	validator, err := NewValidator(config)
	require.NoError(t, err)
	require.NotNil(t, validator)

	assert.Equal(t, "en", validator.defaultLang)
	assert.NotNil(t, validator.validator)
	assert.NotNil(t, validator.uni)
	assert.Len(t, validator.transMap, 2)
}

func TestGetTranslator(t *testing.T) {
	config := Config{
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en", "id"},
	}

	validator, err := NewValidator(config)
	require.NoError(t, err)

	// Test English translator
	enTrans := validator.GetTranslator("en")
	assert.NotNil(t, enTrans)
	assert.Equal(t, "en", enTrans.Locale())

	// Test Indonesian translator
	idTrans := validator.GetTranslator("id")
	assert.NotNil(t, idTrans)
	assert.Equal(t, "id", idTrans.Locale())

	// Test fallback to default
	defaultTrans := validator.GetTranslator("unknown")
	assert.NotNil(t, defaultTrans)
	assert.Equal(t, "en", defaultTrans.Locale())
}

func TestValidateStruct_Success(t *testing.T) {
	config := Config{
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en", "id"},
	}

	validator, err := NewValidator(config)
	require.NoError(t, err)

	type TestStruct struct {
		Name  string `validate:"required,min=2,max=50"`
		Email string `validate:"required,email"`
		Age   int    `validate:"required,min=18,max=100"`
	}

	validData := TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}

	errors := validator.ValidateStruct(validData)
	assert.Nil(t, errors)
}

func TestValidateStruct_ValidationErrors(t *testing.T) {
	config := Config{
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en", "id"},
	}

	validator, err := NewValidator(config)
	require.NoError(t, err)

	type TestStruct struct {
		Name  string `validate:"required,min=2,max=50"`
		Email string `validate:"required,email"`
		Age   int    `validate:"required,min=18,max=100"`
	}

	invalidData := TestStruct{
		Name:  "J",
		Email: "invalid-email",
		Age:   15,
	}

	errors := validator.ValidateStruct(invalidData)
	require.NotNil(t, errors)
	assert.True(t, errors.HasErrors())
	assert.Len(t, errors.Errors, 3)
}

func TestValidate_WithLanguage(t *testing.T) {
	config := Config{
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en", "id"},
	}

	validator, err := NewValidator(config)
	require.NoError(t, err)

	type TestStruct struct {
		Name string `validate:"required"`
	}

	invalidData := TestStruct{
		Name: "",
	}

	// Test English translation
	errorsEN := validator.Validate(invalidData, "en")
	require.NotNil(t, errorsEN)
	assert.Contains(t, errorsEN.Errors[0].Message, "required")

	// Test Indonesian translation
	errorsID := validator.Validate(invalidData, "id")
	require.NotNil(t, errorsID)
	// Indonesian error message should be different
	assert.NotEqual(t, errorsEN.Errors[0].Message, errorsID.Errors[0].Message)
}

func TestValidationErrors_Methods(t *testing.T) {
	errors := &ValidationErrors{
		Errors: []FieldError{
			{
				Field:   "Name",
				Message: "Name is required",
				Tag:     "required",
			},
			{
				Field:   "Email",
				Message: "Email must be a valid email",
				Tag:     "email",
			},
		},
	}

	// Test HasErrors
	assert.True(t, errors.HasErrors())

	// Test GetFirstError
	assert.Equal(t, "Name is required", errors.GetFirstError())

	// Test GetFieldErrors
	fieldErrors := errors.GetFieldErrors()
	assert.Len(t, fieldErrors, 2)
	assert.Equal(t, "Name is required", fieldErrors["Name"])
	assert.Equal(t, "Email must be a valid email", fieldErrors["Email"])

	// Test Error
	assert.Equal(t, "Name is required", errors.Error())

	// Test ToMap
	errorMap := errors.ToMap()
	assert.Contains(t, errorMap, "validation_errors")
}

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"en", "en"},
		{"EN", "en"},
		{"english", "en"},
		{"id", "id"},
		{"ID", "id"},
		{"indonesia", "id"},
		{"indonesian", "id"},
		{"unknown", "en"},
		{"", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeLanguage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
