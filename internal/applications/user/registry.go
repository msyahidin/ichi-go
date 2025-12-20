package user

import (
	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"

	"ichi-go/internal/applications/user/controller"
)

func Register(injector do.Injector, serviceName string, e *echo.Echo) {
	// Register all user domain providers
	RegisterProviders(injector)

	userController := do.MustInvoke[*user.UserController](injector)

	// Register routes
	userController.RegisterRoutes(serviceName, e)
}
