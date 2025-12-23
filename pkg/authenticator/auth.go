package authenticator

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Authenticator struct {
	config     *Config
	jwtAuth    *JWTAuthenticator
	basicAuth  *BasicAuthenticator
	apiKeyAuth *APIKeyAuthenticator
}

func New(config *Config) *Authenticator {
	a := &Authenticator{config: config}

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
}

// // RequirePermission middleware checks ACL
// func (a *Authenticator) RequirePermission(resource, action string) echo.MiddlewareFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
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

func (a *Authenticator) shouldSkip(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

func handleAuthError(c echo.Context, err error) error {
	return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized", err)
}
