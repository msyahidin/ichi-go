package middlewares

import (
	"fmt"
	"ichi-go/pkg/logger"
	"ichi-go/pkg/versioning"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// VersionDeprecation creates a middleware that handles API version deprecation
func VersionDeprecation() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			version := extractVersionFromPath(path)

			if version == "" {
				// No version in path, continue
				return next(c)
			}

			// Parse version
			apiVersion, err := versioning.ParseVersion(version)
			if err != nil {
				// Invalid version format
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"error":   "Invalid API version",
					"message": fmt.Sprintf("Version '%s' is not a valid API version", version),
				})
			}

			// Check deprecation schedule
			if info, exists := versioning.GetDeprecationInfo(apiVersion); exists {
				// Version is sunset - reject the request
				if info.IsSunset() {
					logger.Warnf("Sunset API version accessed: %s from %s", version, c.RealIP())
					return c.JSON(http.StatusGone, map[string]interface{}{
						"error":               "API version sunset",
						"message":             info.Message,
						"deprecated_version":  string(info.Version),
						"replacement_version": string(info.ReplacementVersion),
						"sunset_date":         info.SunsetDate.Format(time.RFC1123),
					})
				}

				// Version is deprecated - add warning headers but allow request
				if info.IsDeprecated() {
					logger.Debugf("Deprecated API version accessed: %s from %s", version, c.RealIP())

					// Add deprecation headers (RFC 8594)
					c.Response().Header().Set("Deprecation", "true")
					c.Response().Header().Set("Sunset", info.SunsetDate.Format(time.RFC1123))

					// Add Link header pointing to replacement version
					replacementPath := strings.Replace(path, string(apiVersion), string(info.ReplacementVersion), 1)
					c.Response().Header().Set("Link", fmt.Sprintf("<%s>; rel=\"successor-version\"", replacementPath))

					// Add custom warning header
					c.Response().Header().Set("X-API-Warning", info.GetWarningMessage())
					c.Response().Header().Set("X-Days-Until-Sunset", fmt.Sprintf("%d", info.DaysUntilSunset()))
				}
			}

			return next(c)
		}
	}
}

// extractVersionFromPath extracts version string from URL path
// Example: /ichi-go/api/v1/auth/login -> "v1"
func extractVersionFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "v") && len(part) > 1 {
			// Check if it looks like a version (v1, v2, etc.)
			if _, err := versioning.ParseVersion(part); err == nil {
				return part
			}
		}
	}
	return ""
}

// VersionValidator creates a middleware that validates API versions against supported versions
func VersionValidator(config *versioning.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !config.Enabled {
				return next(c)
			}

			path := c.Request().URL.Path
			version := extractVersionFromPath(path)

			if version == "" {
				// No version in path, continue
				return next(c)
			}

			// Check if version is supported
			if !config.IsVersionSupported(version) {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"error":              "Unsupported API version",
					"message":            fmt.Sprintf("Version '%s' is not supported", version),
					"supported_versions": config.SupportedVersions,
					"default_version":    config.DefaultVersion,
				})
			}

			// Add version to context for logging
			c.Set("api_version", version)

			return next(c)
		}
	}
}

// VersionLogger creates a middleware that logs API version usage
func VersionLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			version := extractVersionFromPath(path)

			if version != "" {
				logger.Debugf("API Request - Version: %s, Path: %s, Method: %s, IP: %s",
					version,
					path,
					c.Request().Method,
					c.RealIP(),
				)
			}

			return next(c)
		}
	}
}
