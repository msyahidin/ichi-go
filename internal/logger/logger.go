package logger

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"os"
	"rathalos-kit/config"
	"time"
)

var Log zerolog.Logger

func Init() {
	logLevel := viper.GetString("log.level")
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = time.RFC3339

	Log = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(level)
}

func RegisterToMiddleware() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogRequestID: true,
		LogURI:       true,
		LogStatus:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			Log.Info().
				Str("method", c.Request().Method).
				Str("path", c.Request().URL.Path).
				Int("status", c.Response().Status).
				Str("ip", c.RealIP()).
				Str("user_agent", c.Request().UserAgent()).
				Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
				Str("service", config.AppConfig.Name).
				Msg("request completed")
			return nil
		},
	})
}
