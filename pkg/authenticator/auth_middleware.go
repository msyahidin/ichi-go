package authenticator

import (
	"github.com/labstack/echo/v5"
)

// AuthenticateMiddleware Authenticate middleware tries all enabled auth methods
func (a *Authenticator) AuthenticateMiddleware(options ...AuthOption) echo.MiddlewareFunc {
	opts := &authOptions{guestAllowed: false}
	for _, opt := range options {
		opt(opts)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Check skip paths and per-route public endpoints first.
			if a.jwtAuth != nil && a.shouldSkip(c.Path(), a.jwtAuth.config.SkipPaths) {
				return next(c)
			}
			if a.publicEndpoints[c.Request().Method+":"+c.Path()] {
				return next(c)
			}

			// Layer 1: API key — validates the calling client application.
			// Must pass before JWT is evaluated.
			if a.apiKeyAuth != nil && !a.shouldSkip(c.Path(), a.apiKeyAuth.config.SkipPaths) {
				if _, err := a.apiKeyAuth.Authenticate(c); err != nil {
					if opts.guestAllowed {
						return next(c)
					}
					return err
				}
			}

			// Layer 2: JWT — validates the user identity.
			var authCtx *AuthContext
			var authErr error
			if a.jwtAuth != nil {
				authCtx, authErr = a.jwtAuth.Authenticate(c)
			}

			if authCtx == nil {
				if opts.guestAllowed {
					return next(c)
				}
				if authErr != nil {
					return authErr
				}
				return handleAuthError(c, nil)
			}

			// Store auth context for downstream handlers.
			c.Set("auth", authCtx)

			return next(c)
		}
	}
}
