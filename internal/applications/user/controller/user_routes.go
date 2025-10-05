package controller

import (
	"github.com/labstack/echo/v4"
	"ichi-go/internal/middlewares"
)

var domain = "users"

func (c *UserController) httpRoutes(serviceName string, e *echo.Echo) {
	group := e.Group("/" + serviceName + "/api/" + domain)
	group.Use(middlewares.RequestInjector(middlewares.RequestFields{Domain: domain}))
	group.GET("/:id", c.GetUser)
	group.POST("", c.CreateUser)
	group.PUT("/:id", c.UpdateUser)
	group.GET("/pokemon/:name", c.GetPokemon)
}

func (c *UserController) webRoutes(serviceName string, e *echo.Echo) {
	group := e.Group("/" + serviceName + "/" + domain)
	group.GET("/:id", c.GetUserPage)
}

func (c *UserController) RegisterRoutes(serviceName string, e *echo.Echo) {
	c.httpRoutes(serviceName, e)
	c.webRoutes(serviceName, e)
}
