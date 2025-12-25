package server

import (
	"encoding/json"
	"ichi-go/config"
	appConfig "ichi-go/config/app"
	"ichi-go/internal/applications/auth"
	"ichi-go/internal/applications/user"
	"ichi-go/pkg/authenticator"
	"ichi-go/pkg/logger"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"
)

func SetupRestRoutes(injector do.Injector, e *echo.Echo, cfg *config.Config) {
	if err := cfg.Auth().InitializeJWTKeys(); err != nil {
		logger.Errorf("Failed to initialize JWT keys: %v", err)
	}
	authenticator := authenticator.New(cfg.Auth())
	user.Register(injector, cfg.App().Name, e, authenticator)
	auth.Register(injector, cfg.App().Name, e, authenticator)
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
