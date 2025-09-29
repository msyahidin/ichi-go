package user

import (
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"ichi-go/internal/applications/user/controller"
)

func Register(serviceName string, e *echo.Echo, dbConnection *bun.DB, cacheConnection *redis.Client) {
	service := InitializedService(dbConnection, cacheConnection)
	userController := controller.NewUserController(service)
	userController.RegisterRoutes(serviceName, e)
}
