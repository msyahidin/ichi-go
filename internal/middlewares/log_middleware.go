package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"ichi-go/config"
	"ichi-go/pkg/logger"
)

func Logger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogRequestID: true,
		LogURI:       true,
		LogStatus:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Log.Info().
				Str("method", c.Request().Method).
				Str("path", c.Request().URL.Path).
				Int("status", c.Response().Status).
				Str("ip", c.RealIP()).
				Str("user_agent", c.Request().UserAgent()).
				Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
				Str("service", config.Cfg.App.Name).
				Msg("request completed")
			return nil
		},
	})
}
