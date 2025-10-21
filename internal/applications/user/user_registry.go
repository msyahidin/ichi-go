package user

import (
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"ichi-go/internal/applications/user/controller"
	"ichi-go/internal/infra/messaging/rabbitmq"
)

func Register(serviceName string, e *echo.Echo, dbConnection *bun.DB, cacheConnection *redis.Client, mc *rabbitmq.Connection) {
	service := InitializedService(dbConnection, cacheConnection, mc)
	userController := controller.NewUserController(service)
	userController.RegisterRoutes(serviceName, e)
}
