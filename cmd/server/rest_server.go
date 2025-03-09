package server

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"os"
	"rathalos-kit/config"
	"rathalos-kit/internal/applications/user"
)

func SetupRestRoutes(e *echo.Echo) {
	user.Register(GetServiceName(), e)

	// Please register new routes before this line
	if config.AppConfig.Env == "local" {
		generateRouteList(e)
	}
}

func GetServiceName() string {
	return config.AppConfig.Name
}

func generateRouteList(e *echo.Echo) {
	data, err := json.MarshalIndent(e.Routes(), "", "  ")
	if err != nil {
		panic(err)
	}
	os.WriteFile("routes.json", data, 0644)
}
