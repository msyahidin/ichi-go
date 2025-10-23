package controller

import (
	"github.com/labstack/echo/v4"
	"ichi-go/internal/middlewares"
)

var Domain = "users"

func (c *UserController) httpRoutes(serviceName string, e *echo.Echo) {
	group := e.Group("/" + serviceName + "/api/" + Domain)
	group.Use(middlewares.RequestInjector(middlewares.RequestFields{Domain: Domain}))
	group.GET("/:id", c.GetUser)
	group.GET("/:id/send-notification", c.SendNotification)
	group.POST("", c.CreateUser)
	group.PUT("/:id", c.UpdateUser)
	group.GET("/pokemon/:name", c.GetPokemon)
}

func (c *UserController) webRoutes(serviceName string, e *echo.Echo) {
	group := e.Group("/" + serviceName + "/" + Domain)
	group.GET("/:id", c.GetUserPage)
}

func (c *UserController) RegisterRoutes(serviceName string, e *echo.Echo) {
	c.httpRoutes(serviceName, e)
	c.webRoutes(serviceName, e)
}
