package middlewares

import (
	"ichi-go/config"
	"ichi-go/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Logger(config *config.Config) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogRequestID: true,
		LogURI:       true,
		LogStatus:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.RequestLogging(c, "%s %s %d", v.Method, v.URI, v.Status)
			return nil
		},
	})
}
