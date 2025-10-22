package middlewares

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"ichi-go/config"
	httpConfig "ichi-go/config/http"
	"ichi-go/pkg/logger"
	"time"
)

func Init(e *echo.Echo, mainConfig *config.Config) {
	configLog := mainConfig.Log()
	if configLog.RequestIDConfig.Driver == "builtin" {
		e.Use(middleware.RequestID())
	} else {
		e.Use(AppRequestID())
	}
	if configLog.RequestLogging.Enabled {
		switch configLog.RequestLogging.Driver {
		case "builtin":
			e.Use(middleware.Logger())
		case "internal":
			e.Use(Logger(mainConfig))
		default:
			// no request logging
		}
	}

	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogLevel:          log.ERROR,
		DisablePrintStack: !e.Debug,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			logger.Errorf("PANIC RECOVER: %v, stack trace: %s", err, stack)
			return nil
		},
		DisableErrorHandler: true,
	}))

	e.Use(middleware.Gzip())
	e.Use(middleware.Secure())
	e.Use(AppRequestTimeOut(mainConfig.Http()))
	e.Use(Cors(mainConfig.Http().Cors))
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

func AppRequestTimeOut(configHttp httpConfig.HttpConfig) echo.MiddlewareFunc {
	return middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: time.Duration(configHttp.Timeout) * time.Second,
	})
}
