package authenticator

import (
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
