package server

import (
	"github.com/labstack/echo/v4"
	"html/template"
	"net/http"
	"rathalos-kit/config"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w http.ResponseWriter, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func SetupWebRoutes(e *echo.Echo) {
	g := e.Group(config.AppConfig.Name)
	g.GET("", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "<strong>Hello, World!</strong>")
	})
	//templates := template.Must(template.ParseGlob(filepath.Join("web/templates", "*.html")))
	//e.Renderer = &TemplateRenderer{templates}
	//
	//e.Static("/static", "web/static")
	//e.GET("/", func(c echo.Context) error {
	//	data := map[string]interface{}{
	//		"title":   "Home Page",
	//		"message": "Welcome to the Web Server!",
	//	}
	//	return c.Render(http.StatusOK, "index.html", data)
	//})
}
