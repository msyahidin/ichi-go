package authenticator

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

type Authenticator struct {
	config          *Config
	jwtAuth         *JWTAuthenticator
	basicAuth       *BasicAuthenticator
	apiKeyAuth      *APIKeyAuthenticator
	publicEndpoints map[string]bool // "METHOD:full-path" -> true; written at startup only
}

func New(config *Config) *Authenticator {
	a := &Authenticator{
		config:          config,
		publicEndpoints: make(map[string]bool),
	}

	if config.JWT != nil && config.JWT.Enabled {
		a.jwtAuth = NewJWTAuthenticator(config.JWT)
	}
	if config.BasicAuth != nil && config.BasicAuth.Enabled {
		a.basicAuth = NewBasicAuthenticator(config.BasicAuth)
	}
	if config.APIKey != nil && config.APIKey.Enabled {
		a.apiKeyAuth = NewAPIKeyAuthenticator(config.APIKey)
	}
	// if config.ACL != nil && config.ACL.Enabled {
	//     a.aclChecker = NewACLChecker(config.ACL)
	// }

	return a
}

type AuthContext struct {
	UserID UserSubject
	Claims jwt.MapClaims // populated during Authenticate; nil for non-JWT auth
}

// // RequirePermission middleware checks ACL
// func (a *Authenticator) RequirePermission(resource, action string) echo.MiddlewareFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c *echo.Context) error {
// 			authCtx := c.Get("auth").(*AuthContext)

// 			allowed, err := a.aclChecker.Check(authCtx.UserID, resource, action)
// 			if err != nil || !allowed {
// 				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
// 			}

// 			return next(c)
// 		}
// 	}
// }

// Options pattern
type authOptions struct {
	guestAllowed bool
}

type AuthOption func(*authOptions)

func WithGuestAllowed() AuthOption {
	return func(o *authOptions) {
		o.guestAllowed = true
	}
}

// RegisterPublicEndpoint marks a specific method+path combination as public.
// Must be called during server setup before the server starts accepting requests.
func (a *Authenticator) RegisterPublicEndpoint(method, path string) {
	a.publicEndpoints[strings.ToUpper(method)+":"+path] = true
}

func (a *Authenticator) shouldSkip(path string, skipPaths []string) bool {
	for _, p := range skipPaths {
		if strings.HasSuffix(p, "/*") {
			if strings.HasPrefix(path, strings.TrimSuffix(p, "*")) {
				return true
			}
		} else if path == p {
			return true
		}
	}
	return false
}

func handleAuthError(c *echo.Context, err error) error {
	return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized").Wrap(err)
}
