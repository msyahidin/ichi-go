package auth

import (
	"ichi-go/pkg/authenticator"

	authController "ichi-go/internal/applications/auth/controller"

	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"
)

// Register registers all auth domain components
func Register(injector do.Injector, serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
	// Register all auth domain providers
	RegisterProviders(injector)

	// Get auth controller from DI container
	authDI := do.MustInvoke[*authController.AuthController](injector)

	// Register routes
	authDI.RegisterRoutes(serviceName, e, auth)
}
