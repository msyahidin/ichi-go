package auth

import (
	"ichi-go/pkg/authenticator"
	"ichi-go/pkg/versioning"

	"github.com/labstack/echo/v4"
)

var Domain = "auth"

// RegisterRoutes is the main registration method
func (c *AuthController) RegisterRoutes(serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
	c.RegisterRoutesV1(serviceName, e, auth)
	// c.RegisterRoutesV2(serviceName, e, auth)
}

// RegisterRoutesV1 registers V1 auth routes
// This is the current implementation that should be kept as-is
func (c *AuthController) RegisterRoutesV1(serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
	vr := versioning.NewVersionedRoute(serviceName, versioning.TwentySixJan, Domain)

	// Public routes (no authentication required)
	publicGroup := vr.Group(e)
	publicGroup.POST("/login", c.Login)
	publicGroup.POST("/register", c.Register)
	publicGroup.POST("/refresh", c.RefreshToken)

	// Protected routes (authentication required)
	protectedGroup := vr.Group(e)
	protectedGroup.Use(auth.AuthenticateMiddleware())
	protectedGroup.POST("/logout", c.Logout)
	protectedGroup.GET("/me", c.Me)
}

// RegisterRoutesV2 registers V2 auth routes
/*
func (c *AuthController) RegisterRoutesV2(serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
	vr := versioning.NewVersionedRoute(serviceName, versioning.V2, Domain)

	// Public routes
	publicGroup := vr.Group(e)
	publicGroup.POST("/login", c.LoginV2) // New V2 implementation
	publicGroup.POST("/register", c.Register) // Reuse V1
	publicGroup.POST("/refresh", c.RefreshToken) // Reuse V1
	publicGroup.POST("/oauth", c.OAuthLogin) // New V2 endpoint

	// Protected routes
	protectedGroup := vr.Group(e)
	protectedGroup.Use(auth.AuthenticateMiddleware())
	protectedGroup.POST("/logout", c.Logout)
	protectedGroup.GET("/me", c.MeV2) // Enhanced V2 implementation
	protectedGroup.GET("/profile", c.GetProfile) // New V2 endpoint
}
*/
