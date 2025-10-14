package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"ichi-go/config"
	"ichi-go/pkg/logger"
	"net/http"
)

func Logger(config *config.Config) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogRequestID: true,
		LogURI:       true,
		LogStatus:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			getLogger := logger.Log.GetLogger()
			getLogger.Info().
				Str("method", c.Request().Method).
				Str("path", c.Request().URL.Path).
				Int("status", c.Response().Status).
				Str("ip", c.RealIP()).
				Str("user_agent", c.Request().UserAgent()).
				Str(echo.HeaderXRequestID, c.Response().Header().Get(echo.HeaderXRequestID)).
				Str("service", config.App.Name).
				Str("domain", c.Request().Header.Get("domain")).
				Msg(http.StatusText(c.Response().Status))
			return nil
		},
	})
}
