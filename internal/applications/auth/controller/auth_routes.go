package auth

import (
	"ichi-go/pkg/authenticator"

	"github.com/labstack/echo/v4"
)

var Domain = "auth"

// RegisterRoutes registers all authentication routes
func (c *AuthController) RegisterRoutes(serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
	// Public routes (no authentication required)
	authGroup := e.Group("/" + serviceName + "/api/" + Domain)
	authGroup.POST("/login", c.Login)
	authGroup.POST("/register", c.Register)
	authGroup.POST("/refresh", c.RefreshToken)

	// Protected routes (authentication required)
	protectedGroup := e.Group("/" + serviceName + "/api/" + Domain)
	protectedGroup.Use(auth.AuthenticateMiddleware())
	protectedGroup.POST("/logout", c.Logout)
	protectedGroup.GET("/me", c.Me)
}
