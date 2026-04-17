package server

import (
	"html/template"
	"ichi-go/config"
	"io"
	"net/http"

	"github.com/labstack/echo/v5"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(c *echo.Context, w io.Writer, templateName string, data any) error {
	return t.templates.ExecuteTemplate(w, templateName, data)
}

func SetupWebRoutes(e *echo.Echo, config *config.Schema) {
	g := e.Group(config.App.Name)
	g.GET("", func(c *echo.Context) error {
		return c.HTML(http.StatusOK, "<strong>Hello, World!</strong>")
	})
	//templates := template.Must(template.ParseGlob(filepath.Join("web/templates", "*.html")))
	//e.Renderer = &TemplateRenderer{templates}
	//
	//e.Static("/static", "web/static")
	//e.GET("/", func(c *echo.Context) error {
	//	data := map[string]interface{}{
	//		"title":   "Home Page",
	//		"message": "Welcome to the Web Server!",
	//	}
	//	return c.Render(http.StatusOK, "index.html", data)
	//})
}
