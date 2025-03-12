package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"ichi-go/config"
	"time"
)

func Init(e *echo.Echo) {
	if config.Cfg.Log.RequestIDConfig.Driver == "default" {
		e.Use(middleware.RequestID())
	} else {
		e.Use(AppRequestID())
	}
	e.Use(Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.Secure())
	e.Use(AppRequestTimeOut())
	e.Use(Cors())
}

func AppRequestTimeOut() echo.MiddlewareFunc {
	return middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: time.Duration(config.Cfg.Http.Timeout) * time.Second,
	})
}
