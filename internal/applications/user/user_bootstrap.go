package user

import (
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"ichi-go/internal/applications/user/controller"
	"ichi-go/internal/infra/database/ent"
)

func Register(serviceName string, e *echo.Echo, dbConnection *ent.Client, cacheConnection *redis.Client) {
	service := InitializedService(dbConnection, cacheConnection)
	userController := controller.NewUserController(service)
	userController.RegisterRoutes(serviceName, e)
}
