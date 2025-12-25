package user

import (
	"ichi-go/internal/middlewares"
	"ichi-go/pkg/authenticator"

	"github.com/labstack/echo/v4"
)

var Domain = "users"

func (c *UserController) httpRoutes(serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
	group := e.Group("/" + serviceName + "/api/" + Domain)
	group.Use(middlewares.RequestInjector(middlewares.RequestFields{Domain: Domain}))
	group.GET("/:id", c.GetUser)
	group.GET("/:id/send-notification", c.SendNotification)
	group.POST("", c.CreateUser, auth.AuthenticateMiddleware())
	group.PUT("/:id", c.UpdateUser)
	group.GET("/pokemon/:name", c.GetPokemon)
}

func (c *UserController) webRoutes(serviceName string, e *echo.Echo) {
	group := e.Group("/" + serviceName + "/" + Domain)
	group.GET("/:id", c.GetUserPage)
}

func (c *UserController) RegisterRoutes(serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
	c.httpRoutes(serviceName, e, auth)
	c.webRoutes(serviceName, e)
}
