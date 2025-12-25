package user

import (
	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"

	user "ichi-go/internal/applications/user/controller"
	"ichi-go/pkg/authenticator"
)

func Register(injector do.Injector, serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
	// Register all user domain providers
	RegisterProviders(injector)

	userController := do.MustInvoke[*user.UserController](injector)

	// Register routes
	userController.RegisterRoutes(serviceName, e, auth)
}
