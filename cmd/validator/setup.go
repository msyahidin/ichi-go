package validator

import (
	"fmt"

	authValidators "ichi-go/internal/applications/auth/validators"
	appValidator "ichi-go/pkg/validator"
)

// Setup creates and configures validator with all domain validators
func Setup(config appValidator.Config) (*appValidator.AppValidator, error) {
	v, err := appValidator.NewValidator(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	// Register auth validators
	if err := authValidators.RegisterAuthValidators(v); err != nil {
		return nil, fmt.Errorf("failed to register auth validators: %w", err)
	}

	return v, nil
}
