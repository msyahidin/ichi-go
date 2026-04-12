package server

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"regexp"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
	swaggerFiles "github.com/swaggo/files/v2"
	"github.com/swaggo/swag"

	"ichi-go/config"
	"ichi-go/internal/applications/auth"
	healthapp "ichi-go/internal/applications/health"
	notificationapp "ichi-go/internal/applications/notification"
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
	rbacapp.Register(injector, cfg.App().Name, e, appAuth)         // RBAC domain
	notificationapp.Register(injector, cfg.App().Name, e, appAuth) // Notification domain
	healthapp.Register(injector, cfg.App().Name, e, cfg)
}

func GetServiceName(configApp config.AppConfig) string {
	return configApp.Name
}

func generateRouteList(e *echo.Echo) {
	data, err := json.MarshalIndent(e.Router().Routes(), "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("routes.json", data, 0644)
	if err != nil {
		logger.Errorf("failed to write routes to file: %v", err)
	}
}

// swaggerIndexTemplate is the HTML template served at /docs/index.html.
const swaggerIndexTemplate = `<!DOCTYPE html>
<html>
<head>
  <title>Swagger UI</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" type="text/css" href="./swagger-ui.css" >
</head>
<body>
<div id="swagger-ui"></div>
<script src="./swagger-ui-bundle.js"> </script>
<script src="./swagger-ui-standalone-preset.js"> </script>
<script>
window.onload = function() {
  const ui = SwaggerUIBundle({
    url: "doc.json",
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
    plugins: [SwaggerUIBundle.plugins.DownloadUrl],
    layout: "StandaloneLayout"
  })
  window.ui = ui
}
</script>
</body>
</html>`

func openOpenAPIDocs(e *echo.Echo, cfg *config.Config) {
	indexTmpl := template.Must(template.New("swagger_index.html").Parse(swaggerIndexTemplate))
	re := regexp.MustCompile(`^(.*/)([^?].*)?[?|.]*$`)

	handler := func(c *echo.Context) error {
		if c.Request().Method != http.MethodGet {
			return echo.ErrMethodNotAllowed
		}

		matches := re.FindStringSubmatch(c.Request().RequestURI)
		if len(matches) != 3 {
			return echo.ErrNotFound
		}

		path := matches[2]
		switch path {
		case "":
			return c.Redirect(http.StatusMovedPermanently, matches[1]+"index.html")
		case "index.html":
			c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
			return indexTmpl.Execute(c.Response(), nil)
		case "doc.json":
			doc, err := swag.ReadDoc()
			if err != nil {
				return echo.ErrNotFound
			}
			c.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
			_, err = c.Response().Write([]byte(doc))
			return err
		default:
			c.Request().URL.Path = path
			http.FileServer(http.FS(swaggerFiles.FS)).ServeHTTP(c.Response(), c.Request())
			return nil
		}
	}

	e.GET("/docs/*", handler)
	logger.Infof("Swagger UI available at http://localhost:%d/docs/index.html", cfg.Http().Port)
}
