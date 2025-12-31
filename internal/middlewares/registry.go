package middlewares

import (
	"context"
	"fmt"
	"ichi-go/config"
	authValidators "ichi-go/internal/applications/auth/validators"
	httpConfig "ichi-go/pkg/http"
	"ichi-go/pkg/logger"
	appValidator "ichi-go/pkg/validator"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
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

func AppRequestTimeOut(configHttp *httpConfig.Config) echo.MiddlewareFunc {
	return middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: time.Duration(configHttp.Timeout) * time.Second,
	})
}

// setupValidator creates and configures the validator with all domain validators
func setupValidator(e *echo.Echo, config appValidator.Config) error {
	// Create base validator with translation support
	v, err := appValidator.NewValidator(config)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	// Register auth domain validators
	if err := authValidators.RegisterAuthValidators(v); err != nil {
		return fmt.Errorf("failed to register auth validators: %w", err)
	}

	// Register other domain validators here:
	// if err := userValidators.RegisterUserValidators(v); err != nil {
	//     return fmt.Errorf("failed to register user validators: %w", err)
	// }

	// Set Echo validator
	e.Validator = NewValidatorMiddleware(v)

	logger.Debugf("Validator initialized with auth validators")
	return nil
}
