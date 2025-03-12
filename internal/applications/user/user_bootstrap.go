package user

import (
	"github.com/labstack/echo/v4"
	"ichi-go/internal/applications/user/controller"
	"ichi-go/internal/infra/database/ent"
)

func Register(serviceName string, e *echo.Echo, dbConnection *ent.Client) {
	service := InitializedService(dbConnection)
	userController := controller.NewUserController(service)
	userController.RegisterRoutes(serviceName, e)
}
