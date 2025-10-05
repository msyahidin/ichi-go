package middlewares

import (
	"context"
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
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.Secure())
	e.Use(AppRequestTimeOut())
	e.Use(Cors())
	e.Use(copyRequestID)
	e.Use(RequestContextMiddleware())
}

func copyRequestID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := c.Request().Header.Get(echo.HeaderXRequestID)
		if requestID == "" {
			requestID = c.Response().Header().Get(echo.HeaderXRequestID)
		}
		ctx := context.WithValue(c.Request().Context(), echo.HeaderXRequestID, requestID)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

func AppRequestTimeOut() echo.MiddlewareFunc {
	return middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: time.Duration(config.Cfg.Http.Timeout) * time.Second,
	})
}
