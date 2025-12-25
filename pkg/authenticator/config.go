package authenticator

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	JWT       *JWTConfig
	BasicAuth *BasicAuthConfig
	APIKey    *APIKeyConfig
}

func SetDefault() Config {
	return Config{
		JWT: &JWTConfig{
			Enabled:        false,
			SigningMethod:  jwt.SigningMethodRS256,
			TokenLookup:    "header:Authorization",
			AuthScheme:     "Bearer",
			AccessTokenTTL: 15 * time.Minute,
			LeewayDuration: 0,
			PrivateKey:     "sentinel-private.pem",
			PublicKey:      "sentinel-public.pem",
		},
		BasicAuth: &BasicAuthConfig{
			Enabled: false,
			Realm:   "Restricted",
		},
	}
}

// InitializeJWTKeys loads the private and public keys from file paths
// This should be called after creating the config to convert string paths to actual keys
func (c *Config) InitializeJWTKeys() error {
	if c.JWT == nil || !c.JWT.Enabled {
		return nil
	}

	// First, set the signing method from algorithm string
	if err := c.JWT.SetSigningMethod(); err != nil {
		return fmt.Errorf("failed to set signing method: %w", err)
	}

	return c.JWT.InitializeKeys()
}

// SetSigningMethod converts the Algorithm string to jwt.SigningMethod
// This must be called after loading config from YAML/JSON before using the JWT config
func (j *JWTConfig) SetSigningMethod() error {
	if j.Algorithm == "" {
		return fmt.Errorf("algorithm not specified in config")
	}

	switch j.Algorithm {
	case "HS256":
		j.SigningMethod = jwt.SigningMethodHS256
	case "HS384":
		j.SigningMethod = jwt.SigningMethodHS384
	case "HS512":
		j.SigningMethod = jwt.SigningMethodHS512
	case "RS256":
		j.SigningMethod = jwt.SigningMethodRS256
	case "RS384":
		j.SigningMethod = jwt.SigningMethodRS384
	case "RS512":
		j.SigningMethod = jwt.SigningMethodRS512
	case "PS256":
		j.SigningMethod = jwt.SigningMethodPS256
	case "PS384":
		j.SigningMethod = jwt.SigningMethodPS384
	case "PS512":
		j.SigningMethod = jwt.SigningMethodPS512
	case "ES256":
		j.SigningMethod = jwt.SigningMethodES256
	case "ES384":
		j.SigningMethod = jwt.SigningMethodES384
	case "ES512":
		j.SigningMethod = jwt.SigningMethodES512
	default:
		return fmt.Errorf("unsupported algorithm: %s (supported: HS256/384/512, RS256/384/512, PS256/384/512, ES256/384/512)", j.Algorithm)
	}

	// Convert SecretKeyString to SecretKey bytes for HMAC algorithms
	if j.SecretKeyString != "" {
		j.SecretKey = []byte(j.SecretKeyString)
	}

	return nil
}

// InitializeKeys loads keys from file paths configured in JWTConfig
func (j *JWTConfig) InitializeKeys() error {
	// Validate SigningMethod is set
	if j.SigningMethod == nil {
		return fmt.Errorf("SigningMethod is nil - call SetSigningMethod() first or use Config.InitializeJWTKeys()")
	}

	// Skip if keys are already loaded (not string paths)
	if j.PublicKey != nil {
		if _, ok := j.PublicKey.(string); !ok {
			// Key is already loaded
			return nil
		}
	}

	// Load public key from path if it's a string
	if publicKeyPath, ok := j.PublicKey.(string); ok && publicKeyPath != "" {
		publicKey, err := LoadPublicKey(publicKeyPath, j.SigningMethod)
		if err != nil {
			return fmt.Errorf("failed to load public key: %w", err)
		}
		j.PublicKey = publicKey
	}

	// Load private key from path if it's a string
	if privateKeyPath, ok := j.PrivateKey.(string); ok && privateKeyPath != "" {
		privateKey, err := LoadPrivateKey(privateKeyPath, j.SigningMethod)
		if err != nil {
			return fmt.Errorf("failed to load private key: %w", err)
		}
		j.PrivateKey = privateKey
	}

	return nil
}

func (j *JWTConfig) Validate() error {
	if !j.Enabled {
		return nil
	}

	if j.SigningMethod == nil {
		return fmt.Errorf("signing method is required")
	}

	switch j.SigningMethod.(type) {
	case *jwt.SigningMethodHMAC:
		if j.SecretKey == nil || len(j.SecretKey) == 0 {
			return fmt.Errorf("secret key is required for HMAC signing method")
		}

	case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS, *jwt.SigningMethodECDSA:
		if j.PublicKey == nil {
			return fmt.Errorf("public key is required for %s signing method", j.SigningMethod.Alg())
		}
	}

	if j.TokenLookup == "" {
		return fmt.Errorf("token lookup is required")
	}

	if j.AccessTokenTTL <= 0 {
		return fmt.Errorf("access token TTL must be positive")
	}

	if j.RefreshTokenTTL > 0 && j.RefreshTokenTTL <= j.AccessTokenTTL {
		return fmt.Errorf("refresh token TTL should be greater than access token TTL")
	}

	return nil
}
