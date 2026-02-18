package controller

import (
	"github.com/labstack/echo/v4"
	"ichi-go/pkg/authenticator"
)

// RegisterRoutes adds notification API routes to the Echo instance.
//
// Routes:
//   POST /{serviceName}/api/notifications/send  — create and queue a notification campaign
func (c *NotificationController) RegisterRoutes(e *echo.Echo, serviceName string, auth *authenticator.Authenticator) {
	g := e.Group("/" + serviceName + "/api/notifications")
	g.Use(auth.AuthenticateMiddleware())

	g.POST("/send", c.Send)
}
