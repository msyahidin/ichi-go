package user

import (
	"github.com/labstack/echo/v4"
	"rathalos-kit/internal/applications/user/controller"
)

func Register(serviceName string, e *echo.Echo) {
	service := InitializedService()
	userController := controller.NewUserController(service)
	userController.RegisterRoutes(serviceName, e)
}
