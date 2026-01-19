package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"ichi-go/config"
	"ichi-go/pkg/logger"
	"strings"
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
			e.Use(middleware.RequestLogger())
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

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			if strings.Contains(c.Request().URL.Path, "swagger") {
				return true
			}
			return false
		},
	}))
	e.Use(middleware.Secure())
	e.Use(AppRequestTimeOut(mainConfig.Http()))
	e.Use(Cors(&mainConfig.Http().Cors))
	e.Use(copyRequestID)
	e.Use(RequestContextMiddleware())

	versionConfig := mainConfig.Versioning()
	if versionConfig != nil && versionConfig.Enabled {
		// Log API version usage
		e.Use(VersionLogger())

		// Validate version is supported
		e.Use(VersionValidator(versionConfig))

		// Check deprecation and add warning headers
		if versionConfig.Deprecation.HeaderEnabled {
			e.Use(VersionDeprecation())
		}

		logger.Infof("API versioning enabled - Default: %s, Supported: %v",
			versionConfig.DefaultVersion,
			versionConfig.SupportedVersions)
	}

	// Initialize validator
	if err := setupValidator(e, *mainConfig.Validator()); err != nil {
		logger.Fatalf("failed to initialize validator: %v", err)
	}
}
