package controller

import "github.com/labstack/echo/v4"

var domain = "users"

func (c *UserController) httpRoutes(serviceName string, e *echo.Echo) {
	group := e.Group("/" + serviceName + "/api/" + domain)
	group.GET("/:id", c.GetUser)
}

func (c *UserController) webRoutes(serviceName string, e *echo.Echo) {
	group := e.Group("/" + serviceName + "/" + domain)
	group.GET("/:id", c.GetUserPage)
}

func (c *UserController) RegisterRoutes(serviceName string, e *echo.Echo) {
	c.httpRoutes(serviceName, e)
	c.webRoutes(serviceName, e)
}
