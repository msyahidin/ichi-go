package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"rathalos-kit/config"
	"rathalos-kit/internal/logger"
	"time"
)

func Init(e *echo.Echo) {
	if config.AppConfig.Log.RequestIDConfig.Driver == "default" {
		e.Use(middleware.RequestID())
	} else {
		e.Use(AppRequestID())
	}
	e.Use(logger.RegisterToMiddleware())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.Secure())
	e.Use(AppRequestTimeOut())
}

func AppRequestTimeOut() echo.MiddlewareFunc {
	return middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: time.Duration(config.AppConfig.Request.Timeout),
	})
}
