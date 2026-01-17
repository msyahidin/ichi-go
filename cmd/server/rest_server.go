package server

import (
	"encoding/json"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"
	echoSwagger "github.com/swaggo/echo-swagger"

	"ichi-go/config"
	"ichi-go/internal/applications/auth"
	healthapp "ichi-go/internal/applications/health"
	rbacapp "ichi-go/internal/applications/rbac"
	"ichi-go/internal/applications/user"
	"ichi-go/pkg/authenticator"
	"ichi-go/pkg/logger"
)

func SetupRestRoutes(injector do.Injector, e *echo.Echo, cfg *config.Config) {
	openOpenAPIDocs(e, cfg)
	if err := cfg.Auth().InitializeJWTKeys(); err != nil {
		logger.Errorf("Failed to initialize JWT keys: %v", err)
	}
	appAuth := authenticator.New(cfg.Auth())

	// Register application domains
	user.Register(injector, cfg.App().Name, e, appAuth)
	auth.Register(injector, cfg.App().Name, e, appAuth)
	rbacapp.Register(injector, cfg.App().Name, e, appAuth) // RBAC domain
	healthapp.Register(injector, cfg.App().Name, e, cfg)
}

func GetServiceName(configApp config.AppConfig) string {
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

func openOpenAPIDocs(e *echo.Echo, cfg *config.Config) {
	// Swagger documentation endpoint
	e.GET("/docs/*", echoSwagger.WrapHandler)
	logger.Infof("Swagger UI available at http://localhost:%d/docs/index.html", cfg.Http().Port)
}
