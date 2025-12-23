package authenticator

import (
	"ichi-go/pkg/utils/response"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Authenticate middleware tries all enabled auth methods
func (a *Authenticator) AuthenticateMiddleware(options ...AuthOption) echo.MiddlewareFunc {
	opts := &authOptions{guestAllowed: false}
	for _, opt := range options {
		opt(opts)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			// Try each auth method
			var authCtx *AuthContext
			var err error

			// Priority: JWT > API Key > Basic Auth > OAuth2
			if a.jwtAuth != nil {
				if a.shouldSkip(c.Path(), a.jwtAuth.config.SkipPaths) {
					return next(c)
				}
				authCtx, err = a.jwtAuth.Authenticate(c)
			}
			// if authCtx == nil && a.apiKeyAuth != nil {
			// 	authCtx, err = a.apiKeyAuth.Authenticate(c)
			// }
			// if authCtx == nil && a.basicAuth != nil {
			// 	authCtx, err = a.basicAuth.Authenticate(c)
			// }

			// Handle authentication failure
			if authCtx == nil {
				if opts.guestAllowed {
					return next(c)
				}
				return response.Error(c, http.StatusUnauthorized, err)
			}

			// Store auth context
			c.Set("auth", authCtx)

			return next(c)
		}
	}
}
