package authenticator

import (
	pkgErrors "ichi-go/pkg/errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type JWTAuthenticator struct {
	config *JWTConfig
}

func NewJWTAuthenticator(config *JWTConfig) *JWTAuthenticator {
	return &JWTAuthenticator{config: config}
}

// JWTConfig represents the configuration for JWT authentication
type JWTConfig struct {
	Enabled bool

	Algorithm string `yaml:"signing_method" json:"signing_method" mapstructure:"signing_method"`

	SecretKeyString string `yaml:"secret_key" json:"secret_key" mapstructure:"secret_key"`

	SigningMethod jwt.SigningMethod `yaml:"-" json:"-"`
	SecretKey     []byte            `yaml:"-" json:"-"`
	PublicKey     interface{}       `yaml:"public_key" json:"public_key" mapstructure:"public_key"`
	PrivateKey    interface{}       `yaml:"private_key" json:"private_key" mapstructure:"private_key"`

	TokenLookup string `yaml:"token_lookup" json:"token_lookup" mapstructure:"token_lookup"`
	AuthScheme  string `yaml:"auth_scheme" json:"auth_scheme" mapstructure:"auth_scheme"`

	Issuer        string   `yaml:"issuer" json:"issuer" mapstructure:"issuer"`
	Audience      []string `yaml:"audience" json:"audience" mapstructure:"audience"`
	RequireClaims []string `yaml:"require_claims" json:"require_claims" mapstructure:"require_claims"`

	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" json:"access_token_ttl" mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" json:"refresh_token_ttl" mapstructure:"refresh_token_ttl"`

	ClaimsValidator JWTClaimsValidator `yaml:"-" json:"-"`
	KeyFunc         jwt.Keyfunc        `yaml:"-" json:"-"`
	SkipPaths       []string           `yaml:"skip_paths" json:"skip_paths" mapstructure:"skip_paths"`

	LeewayDuration time.Duration `yaml:"leeway_duration" json:"leeway_duration" mapstructure:"leeway_duration"`
}

type JWTClaimsValidator func(claims jwt.MapClaims) error

func (a *JWTAuthenticator) Authenticate(c echo.Context) (*AuthContext, error) {
	tokenString, err := ExtractToken(c, *a.config)
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeInvalidToken).
			Hint("Token not found in request").
			Wrap(err)
	}

	token, err := ParseToken(tokenString, *a.config)
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeInvalidToken).
			Hint("Invalid token format").
			Wrap(err)
	}

	claims, err := ValidateToken(token, *a.config)
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeTokenExpired).
			Hint("Token validation failed").
			Wrap(err)
	}

	user, err := GetUserIdFromMapClaims(claims)
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeInvalidToken).
			Hint("Invalid user claims").
			Wrap(err)
	}

	return &AuthContext{UserID: user}, nil
}

// GenerateToken creates a new JWT token for the given user ID
// This is a convenience method for the authenticator
func (a *JWTAuthenticator) GenerateToken(userID uint64) (string, error) {
	return GenerateAccessToken(userID, *a.config)
}

// GenerateTokens creates both access and refresh tokens for the given user ID
// Returns a TokenPair with both tokens
func (a *JWTAuthenticator) GenerateTokens(userID uint64) (*TokenPair, error) {
	return GenerateTokenPair(userID, *a.config)
}

// ValidateRefreshToken validates a refresh token and returns the user ID
// This is useful for token refresh endpoints
func (a *JWTAuthenticator) ValidateRefreshToken(tokenString string) (uint64, error) {
	token, err := ParseToken(tokenString, *a.config)
	if err != nil {
		return 0, pkgErrors.AuthService(pkgErrors.ErrCodeInvalidToken).
			Hint("Invalid refresh token format").
			Wrap(err)
	}

	claims, err := ValidateToken(token, *a.config)
	if err != nil {
		return 0, pkgErrors.AuthService(pkgErrors.ErrCodeTokenExpired).
			Hint("Refresh token expired or invalid").
			Wrap(err)
	}

	user, err := GetUserIdFromMapClaims(claims)
	if err != nil {
		return 0, pkgErrors.AuthService(pkgErrors.ErrCodeInvalidToken).
			Hint("Invalid user claims in refresh token").
			Wrap(err)
	}

	return user.ID, nil
}
