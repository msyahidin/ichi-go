package server

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"os"
	"rathalos-kit/config"
	"rathalos-kit/internal/applications/user"
	"rathalos-kit/internal/infrastructure/database/ent"
)

func SetupRestRoutes(e *echo.Echo, dbClient *ent.Client) {
	user.Register(GetServiceName(), e, dbClient)

	// Please register new domain routes before this line
	if config.Cfg.App.Env == "local" {
		generateRouteList(e)
	}
}

func GetServiceName() string {
	return config.Cfg.App.Name
}

func generateRouteList(e *echo.Echo) {
	data, err := json.MarshalIndent(e.Routes(), "", "  ")
	if err != nil {
		panic(err)
	}
	os.WriteFile("routes.json", data, 0644)
}
