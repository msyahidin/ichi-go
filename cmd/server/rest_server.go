package server

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"ichi-go/config"
	"ichi-go/internal/applications/user"
	"os"
)

func SetupRestRoutes(e *echo.Echo, dbClient *bun.DB, cacheClient *redis.Client) {
	user.Register(GetServiceName(), e, dbClient, cacheClient)

	// Please register new domain routes before this line
	if config.App().Env == "local" {
		generateRouteList(e)
	}
}

func GetServiceName() string {
	return config.App().Name
}

func generateRouteList(e *echo.Echo) {
	data, err := json.MarshalIndent(e.Routes(), "", "  ")
	if err != nil {
		panic(err)
	}
	os.WriteFile("routes.json", data, 0644)
}
