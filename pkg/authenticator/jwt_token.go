package authenticator

import (
	"fmt"
	"ichi-go/pkg/requestctx"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// ExtractToken extracts JWT token from Echo context based on TokenLookup configuration
// Supports multiple lookup strategies: header, query, cookie
func ExtractToken(c echo.Context, config JWTConfig) (string, error) {
	// Parse TokenLookup string (e.g., "header:Authorization,query:token,cookie:jwt")
	lookups := parseTokenLookup(config.TokenLookup)

	for _, lookup := range lookups {
		token := ""
		switch lookup.source {
		case "header":
			token = c.Request().Header.Get(lookup.key)
		case "query":
			token = c.QueryParam(lookup.key)
		case "cookie":
			cookie, err := c.Cookie(lookup.key)
			if err == nil && cookie != nil {
				token = cookie.Value
			}
		}

		if token != "" {
			// Strip auth scheme if present (e.g., "Bearer ")
			if config.AuthScheme != "" {
				prefix := config.AuthScheme + " "
				if strings.HasPrefix(token, prefix) {
					token = strings.TrimPrefix(token, prefix)
				}
			}
			return token, nil
		}
	}

	return "", jwt.ErrTokenMalformed
}

// ExtractTokenFromContext extracts JWT token from RequestContext
// This is a convenience function for when you already have RequestContext
func ExtractTokenFromContext(reqCtx requestctx.RequestContext, config JWTConfig) (string, error) {
	authHeader := reqCtx.Authorization
	if authHeader == "" {
		return "", jwt.ErrTokenMalformed
	}

	// Strip auth scheme
	if config.AuthScheme != "" {
		prefix := config.AuthScheme + " "
		if strings.HasPrefix(authHeader, prefix) {
			return strings.TrimPrefix(authHeader, prefix), nil
		}
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return "", jwt.ErrTokenMalformed
	}

	return parts[1], nil
}

// ParseToken parses a JWT token string and validates its signature
// Supports multiple signing methods: HMAC (HS*), RSA (RS*), ECDSA (ES*)
func ParseToken(tokenString string, config JWTConfig) (*jwt.Token, error) {
	// Set up parser options
	var parserOptions []jwt.ParserOption

	// Add leeway for time-based claims if configured
	if config.LeewayDuration > 0 {
		parserOptions = append(parserOptions, jwt.WithLeeway(config.LeewayDuration))
	}

	// Add issuer validation if configured
	if config.Issuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(config.Issuer))
	}

	// // Add audience validation if configured
	// if len(config.Audience) > 0 {
	// 	parserOptions = append(parserOptions, jwt.WithAudience(config.Audience...))
	// }

	// Parse token with key function
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Use custom KeyFunc if provided (for multi-tenant scenarios)
		if config.KeyFunc != nil {
			return config.KeyFunc(token)
		}

		// Verify signing method matches configuration
		if !isSigningMethodValid(token.Method, config.SigningMethod) {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
		}

		// Return appropriate key based on signing method
		switch config.SigningMethod.(type) {
		case *jwt.SigningMethodHMAC:
			return config.SecretKey, nil
		case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS, *jwt.SigningMethodECDSA:
			return config.PublicKey, nil
		default:
			return nil, fmt.Errorf("unsupported signing method: %s", config.SigningMethod.Alg())
		}
	}, parserOptions...)

	if err != nil {
		return nil, err
	}

	if token == nil {
		return nil, jwt.ErrTokenMalformed
	}

	return token, nil
}

// ValidateToken validates JWT token and extracts claims
// Performs comprehensive validation including custom claims validation
func ValidateToken(token *jwt.Token, config JWTConfig) (jwt.MapClaims, error) {
	// Check if token is valid
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	// Validate required claims are present
	if err := validateRequiredClaims(claims, config.RequireClaims); err != nil {
		return nil, err
	}

	// Run custom claims validator if provided
	if config.ClaimsValidator != nil {
		if err := config.ClaimsValidator(claims); err != nil {
			return nil, fmt.Errorf("custom claims validation failed: %w", err)
		}
	}

	return claims, nil
}

// parseTokenLookup parses the TokenLookup configuration string
// Format: "source1:key1,source2:key2" (e.g., "header:Authorization,query:token")
func parseTokenLookup(lookupStr string) []tokenLookup {
	if lookupStr == "" {
		// Default to Authorization header
		return []tokenLookup{{source: "header", key: "Authorization"}}
	}

	var lookups []tokenLookup
	parts := strings.Split(lookupStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		sourceParts := strings.Split(part, ":")
		if len(sourceParts) != 2 {
			continue
		}

		lookups = append(lookups, tokenLookup{
			source: strings.TrimSpace(sourceParts[0]),
			key:    strings.TrimSpace(sourceParts[1]),
		})
	}

	return lookups
}

// validateRequiredClaims checks if all required claims are present in the token
func validateRequiredClaims(claims jwt.MapClaims, requiredClaims []string) error {
	for _, required := range requiredClaims {
		if _, exists := claims[required]; !exists {
			return fmt.Errorf("required claim '%s' is missing", required)
		}
	}
	return nil
}

// isSigningMethodValid checks if the token's signing method matches the expected method
func isSigningMethodValid(tokenMethod, expectedMethod jwt.SigningMethod) bool {
	// Check exact match
	if tokenMethod.Alg() == expectedMethod.Alg() {
		return true
	}

	// Check method family match (e.g., both are HMAC, both are RSA, etc.)
	switch expectedMethod.(type) {
	case *jwt.SigningMethodHMAC:
		_, ok := tokenMethod.(*jwt.SigningMethodHMAC)
		return ok
	case *jwt.SigningMethodRSA:
		_, ok := tokenMethod.(*jwt.SigningMethodRSA)
		return ok
	case *jwt.SigningMethodRSAPSS:
		_, ok := tokenMethod.(*jwt.SigningMethodRSAPSS)
		return ok
	case *jwt.SigningMethodECDSA:
		_, ok := tokenMethod.(*jwt.SigningMethodECDSA)
		return ok
	}

	return false
}

// tokenLookup represents a token lookup configuration
type tokenLookup struct {
	source string // "header", "query", or "cookie"
	key    string // the name of the header/query/cookie
}
