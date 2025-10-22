package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"ichi-go/config"
	"ichi-go/pkg/logger"
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
