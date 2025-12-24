package authenticator

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// StandardClaimsInput contains the input parameters for creating standard JWT claims
type StandardClaimsInput struct {
	UserID       uint64                 // Subject (user identifier)
	Issuer       string                 // Token issuer (iss claim)
	Audience     []string               // Token audience (aud claim)
	ExpiresIn    time.Duration          // Token expiration duration from now
	NotBefore    *time.Time             // Token not valid before this time (optional)
	IssuedAt     *time.Time             // Token issued at time (optional, defaults to now)
	JWTID        string                 // Unique token identifier (optional)
	CustomClaims map[string]interface{} // Additional custom claims
}

// CreateStandardClaims creates JWT RegisteredClaims with standard fields
// This is useful when you want to use jwt.RegisteredClaims type
func CreateStandardClaims(input StandardClaimsInput) jwt.RegisteredClaims {
	now := time.Now()

	// Use provided IssuedAt or default to now
	issuedAt := now
	if input.IssuedAt != nil {
		issuedAt = *input.IssuedAt
	}

	claims := jwt.RegisteredClaims{
		Issuer:    input.Issuer,
		Subject:   strconv.FormatUint(input.UserID, 10),
		Audience:  input.Audience,
		ExpiresAt: jwt.NewNumericDate(now.Add(input.ExpiresIn)),
		IssuedAt:  jwt.NewNumericDate(issuedAt),
	}

	// Add optional NotBefore
	if input.NotBefore != nil {
		claims.NotBefore = jwt.NewNumericDate(*input.NotBefore)
	}

	// Add optional JWT ID
	if input.JWTID != "" {
		claims.ID = input.JWTID
	}

	return claims
}

// CreateCustomClaims creates JWT MapClaims with both standard and custom fields
// This is the most flexible approach and recommended for most use cases
func CreateCustomClaims(input StandardClaimsInput) jwt.MapClaims {
	now := time.Now()

	// Use provided IssuedAt or default to now
	issuedAt := now
	if input.IssuedAt != nil {
		issuedAt = *input.IssuedAt
	}

	claims := jwt.MapClaims{
		"sub": strconv.FormatUint(input.UserID, 10),
		"iat": issuedAt.Unix(),
		"exp": now.Add(input.ExpiresIn).Unix(),
	}

	// Add standard claims if provided
	if input.Issuer != "" {
		claims["iss"] = input.Issuer
	}

	if len(input.Audience) > 0 {
		if len(input.Audience) == 1 {
			claims["aud"] = input.Audience[0]
		} else {
			claims["aud"] = input.Audience
		}
	}

	if input.NotBefore != nil {
		claims["nbf"] = input.NotBefore.Unix()
	}

	if input.JWTID != "" {
		claims["jti"] = input.JWTID
	}

	// Merge custom claims
	if input.CustomClaims != nil {
		for key, value := range input.CustomClaims {
			// Don't overwrite standard claims
			if key != "sub" && key != "iat" && key != "exp" &&
				key != "iss" && key != "aud" && key != "nbf" && key != "jti" {
				claims[key] = value
			}
		}
	}

	return claims
}

// GenerateToken creates and signs a JWT token with the provided claims
// Works with both jwt.RegisteredClaims and jwt.MapClaims
func GenerateToken(claims jwt.Claims, config JWTConfig) (string, error) {
	// Create token with claims
	token := jwt.NewWithClaims(config.SigningMethod, claims)

	// Sign token based on signing method
	var signedToken string
	var err error

	switch config.SigningMethod.(type) {
	case *jwt.SigningMethodHMAC:
		if config.SecretKey == nil {
			return "", fmt.Errorf("secret key is required for HMAC signing")
		}
		signedToken, err = token.SignedString(config.SecretKey)

	case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS:
		if config.PrivateKey == nil {
			return "", fmt.Errorf("private key is required for RSA signing")
		}
		signedToken, err = token.SignedString(config.PrivateKey)

	case *jwt.SigningMethodECDSA:
		if config.PrivateKey == nil {
			return "", fmt.Errorf("private key is required for ECDSA signing")
		}
		signedToken, err = token.SignedString(config.PrivateKey)

	default:
		return "", fmt.Errorf("unsupported signing method: %s", config.SigningMethod.Alg())
	}

	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// GenerateAccessToken is a convenience function to generate an access token
// Uses the AccessTokenTTL from config
func GenerateAccessToken(userID uint64, config JWTConfig) (string, error) {
	input := StandardClaimsInput{
		UserID:    userID,
		Issuer:    config.Issuer,
		Audience:  config.Audience,
		ExpiresIn: config.AccessTokenTTL,
	}

	claims := CreateCustomClaims(input)
	return GenerateToken(claims, config)
}

// GenerateRefreshToken is a convenience function to generate a refresh token
// Uses the RefreshTokenTTL from config
func GenerateRefreshToken(userID uint64, config JWTConfig) (string, error) {
	input := StandardClaimsInput{
		UserID:    userID,
		Issuer:    config.Issuer,
		Audience:  config.Audience,
		ExpiresIn: config.RefreshTokenTTL,
	}

	claims := CreateCustomClaims(input)
	return GenerateToken(claims, config)
}

// TokenPair represents an access token and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until access token expires
}

// GenerateTokenPair generates both access and refresh tokens
// This is the recommended way to generate tokens for authentication flows
func GenerateTokenPair(userID uint64, config JWTConfig) (*TokenPair, error) {
	// Generate access token
	accessToken, err := GenerateAccessToken(userID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := GenerateRefreshToken(userID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    config.AuthScheme,
		ExpiresIn:    int64(config.AccessTokenTTL.Seconds()),
	}, nil
}
