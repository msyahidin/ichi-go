package authenticator

import (
	"crypto/subtle"
	"errors"

	pkgErrors "ichi-go/pkg/errors"

	"github.com/labstack/echo/v5"
)

const defaultAPIKeyHeader = "X-API-Key"

func NewAPIKeyAuthenticator(config *APIKeyConfig) *APIKeyAuthenticator {
	if config.Header == "" {
		config.Header = defaultAPIKeyHeader
	}
	return &APIKeyAuthenticator{config: config}
}

type APIKeyAuthenticator struct {
	config *APIKeyConfig
}

type APIKeyConfig struct {
	Enabled   bool            `yaml:"enabled"    mapstructure:"enabled"`
	Header    string          `yaml:"header"     mapstructure:"header"`
	Keys      []string        `yaml:"keys"       mapstructure:"keys"`
	SkipPaths []string        `yaml:"skip_paths" mapstructure:"skip_paths"`
	Validator APIKeyValidator `yaml:"-"          mapstructure:"-"`
}

// APIKeyValidator is an optional custom validator. Return an error to reject the key.
// When nil, Authenticate falls back to the static Keys list.
type APIKeyValidator func(apiKey string) error

func (a *APIKeyAuthenticator) Authenticate(c *echo.Context) (*AuthContext, error) {
	key := c.Request().Header.Get(a.config.Header)
	if key == "" {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeUnauthorized).
			Hint("missing " + a.config.Header + " header").
			Wrap(errors.New("missing API key"))
	}

	if a.config.Validator != nil {
		if err := a.config.Validator(key); err != nil {
			return nil, pkgErrors.AuthService(pkgErrors.ErrCodeInvalidToken).
				Hint("API key rejected by validator").
				Wrap(err)
		}
		return &AuthContext{}, nil
	}

	if len(a.config.Keys) == 0 {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeUnauthorized).
			Hint("no static keys configured and no Validator set").
			Wrap(errors.New("API key authenticator misconfigured"))
	}

	// Compare against every key without short-circuiting to avoid leaking
	// which position in the list matched (timing side-channel).
	keyBytes := []byte(key)
	var matched int
	for _, valid := range a.config.Keys {
		matched |= subtle.ConstantTimeCompare(keyBytes, []byte(valid))
	}
	if matched == 1 {
		return &AuthContext{}, nil
	}

	return nil, pkgErrors.AuthService(pkgErrors.ErrCodeInvalidToken).
		Hint("invalid API key").
		Wrap(errors.New("API key not recognised"))
}
