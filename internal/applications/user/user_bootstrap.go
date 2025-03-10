package user

import (
	"github.com/labstack/echo/v4"
	"rathalos-kit/internal/applications/user/controller"
	"rathalos-kit/internal/infrastructure/database/ent"
)

func Register(serviceName string, e *echo.Echo, dbConnection *ent.Client) {
	service := InitializedService(dbConnection)
	userController := controller.NewUserController(service)
	userController.RegisterRoutes(serviceName, e)
}
