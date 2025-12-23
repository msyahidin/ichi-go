package authenticator

import (
	"ichi-go/pkg/requestctx"
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

type JWTConfig struct {
	Enabled bool

	// Token validation
	SigningMethod jwt.SigningMethod // e.g., jwt.SigningMethodHS256, RS256
	SecretKey     []byte            // For HMAC (HS256, HS384, HS512)
	PublicKey     interface{}       // For RSA/ECDSA (RS256, ES256, etc.)
	PrivateKey    interface{}       // For token generation

	// Token extraction
	TokenLookup string // "header:Authorization,query:token,cookie:jwt"
	AuthScheme  string // "Bearer" (default) or custom

	// Token validation rules
	Issuer        string   // Expected "iss" claim
	Audience      []string // Expected "aud" claim
	RequireClaims []string // Required claim keys (e.g., ["sub", "exp"])

	// Token lifecycle
	AccessTokenTTL  time.Duration // e.g., 15 * time.Minute
	RefreshTokenTTL time.Duration // e.g., 7 * 24 * time.Hour

	// Advanced
	ClaimsValidator JWTClaimsValidator // Custom claims validation
	KeyFunc         jwt.Keyfunc        // Dynamic key resolution for multi-tenant
	SkipPaths       []string           // JWT-specific skip paths

	LeewayDuration time.Duration // Clock skew leeway for time-based claims
}

type JWTClaimsValidator func(claims jwt.MapClaims) error

func (a *JWTAuthenticator) Authenticate(c echo.Context) (*AuthContext, error) {
	// Extract token from request
	reqCtx := requestctx.FromContext(c.Request().Context())
	tokenString, err := ExtractToken(*reqCtx)
	if err != nil {
		return nil, err
	}

	// Parse token
	token, err := ParseToken(tokenString, *a.config)
	if err != nil {
		return nil, err
	}

	// Validate token
	claims, err := ValidateToken(*token)
	if err != nil {
		return nil, err
	}

	// Get user ID
	user, err := GetUserId(claims)
	if err != nil {
		return nil, err
	}

	return &AuthContext{UserID: user}, nil
}
