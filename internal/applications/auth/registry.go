package auth

import (
	"ichi-go/pkg/authenticator"

	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"
	authController "ichi-go/internal/applications/auth/controller"
)

// Register registers all auth domain components
func Register(injector do.Injector, serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
	// Register all auth domain providers
	RegisterProviders(injector)

	// Get auth controller from DI container
	authController := do.MustInvoke[*authController.AuthController](injector)

	// Register routes
	authController.RegisterRoutes(serviceName, e, auth)
}
