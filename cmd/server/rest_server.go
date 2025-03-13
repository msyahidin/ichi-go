package server

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"ichi-go/config"
	"ichi-go/internal/applications/user"
	"ichi-go/internal/infra/database/ent"
	"os"
)

func SetupRestRoutes(e *echo.Echo, dbClient *ent.Client) {
	user.Register(GetServiceName(), e, dbClient)

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
