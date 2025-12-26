package validators

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestValidateStrongPassword(t *testing.T) {
	v := validator.New()
	v.RegisterValidation("strong_password", validateStrongPassword)

	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{
			name:     "valid strong password",
			password: "Pass123!@#",
			valid:    true,
		},
		{
			name:     "valid with special chars",
			password: "MyP@ssw0rd",
			valid:    true,
		},
		{
			name:     "too short",
			password: "Pass1!",
			valid:    false,
		},
		{
			name:     "no uppercase",
			password: "pass123!@#",
			valid:    false,
		},
		{
			name:     "no lowercase",
			password: "PASS123!@#",
			valid:    false,
		},
		{
			name:     "no number",
			password: "Password!@#",
			valid:    false,
		},
		{
			name:     "no special char",
			password: "Password123",
			valid:    false,
		},
		{
			name:     "empty",
			password: "",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Var(tt.password, "strong_password")
			isValid := err == nil
			assert.Equal(t, tt.valid, isValid, "password: %s", tt.password)
		})
	}
}

func TestValidateUsernameFormat(t *testing.T) {
	v := validator.New()
	v.RegisterValidation("username_format", validateUsernameFormat)

	tests := []struct {
		name     string
		username string
		valid    bool
	}{
		{
			name:     "valid simple username",
			username: "johndoe",
			valid:    true,
		},
		{
			name:     "valid with numbers",
			username: "user123",
			valid:    true,
		},
		{
			name:     "valid with underscore",
			username: "john_doe",
			valid:    true,
		},
		{
			name:     "valid with hyphen",
			username: "john-doe",
			valid:    true,
		},
		{
			name:     "valid mixed",
			username: "User_123-test",
			valid:    true,
		},
		{
			name:     "starts with number",
			username: "123user",
			valid:    false,
		},
		{
			name:     "starts with underscore",
			username: "_username",
			valid:    false,
		},
		{
			name:     "starts with hyphen",
			username: "-username",
			valid:    false,
		},
		{
			name:     "contains space",
			username: "john doe",
			valid:    false,
		},
		{
			name:     "contains special chars",
			username: "john@doe",
			valid:    false,
		},
		{
			name:     "too short",
			username: "ab",
			valid:    false,
		},
		{
			name:     "too long",
			username: "a" + string(make([]byte, 60)),
			valid:    false,
		},
		{
			name:     "empty",
			username: "",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Var(tt.username, "username_format")
			isValid := err == nil
			assert.Equal(t, tt.valid, isValid, "username: %s", tt.username)
		})
	}
}

func TestValidateUsernameFormat_MinMaxLength(t *testing.T) {
	v := validator.New()
	v.RegisterValidation("username_format", validateUsernameFormat)

	// Test minimum length (3)
	assert.NoError(t, v.Var("abc", "username_format"))
	assert.Error(t, v.Var("ab", "username_format"))

	// Test maximum length (50)
	validMax := "a" + string(make([]byte, 49))
	assert.NoError(t, v.Var(validMax, "username_format"))

	invalidMax := "a" + string(make([]byte, 50))
	assert.Error(t, v.Var(invalidMax, "username_format"))
}
