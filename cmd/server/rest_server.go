package server

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"ichi-go/config"
	appConfig "ichi-go/config/app"
	"ichi-go/internal/applications/user"
	"ichi-go/internal/infra/database/ent"
	"os"
)

func SetupRestRoutes(e *echo.Echo, config *config.Config, dbClient *ent.Client, cacheClient *redis.Client) {
	user.Register(GetServiceName(&config.App), e, dbClient, cacheClient)

	// Please register new domain routes before this line
	if e.Debug {
		generateRouteList(e)
	}
}

func GetServiceName(configApp *appConfig.AppConfig) string {
	return configApp.Name
}

func generateRouteList(e *echo.Echo) {
	data, err := json.MarshalIndent(e.Routes(), "", "  ")
	if err != nil {
		panic(err)
	}
	os.WriteFile("routes.json", data, 0644)
}
