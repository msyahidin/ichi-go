package server

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"
	"ichi-go/config"
	appConfig "ichi-go/config/app"
	"ichi-go/internal/applications/user"
	"ichi-go/pkg/logger"
	"os"
)

//func SetupRestRoutes(e *echo.Echo, mainConfig *config.Config, dbClient *bun.DB, cacheClient *redis.Client, messagingConnection *rabbitmq.Connection) {
//	//user_wire.Register(mainConfig.App().Name, e, dbClient, cacheClient, messagingConnection)
//	// Please register new domain routes before this line
//	if mainConfig.App().Env == "local" {
//		generateRouteList(e)
//	}
//}

func SetupRestRoutes(injector do.Injector, e *echo.Echo, cfg *config.Config) {
	user.Register(injector, cfg.App().Name, e)
}

func GetServiceName(configApp appConfig.AppConfig) string {
	return configApp.Name
}

func generateRouteList(e *echo.Echo) {
	data, err := json.MarshalIndent(e.Routes(), "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("routes.json", data, 0644)
	if err != nil {
		logger.Errorf("failed to write routes to file: %v", err)
	}
}
