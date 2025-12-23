package middlewares

import (
	httpConfig "ichi-go/config/http"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func GetCorsConfig(configCors *httpConfig.CorsConfig) middleware.CORSConfig {
	allowedOrigins := configCors.AllowOrigins
	return middleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
	}
}

func Cors(configCors *httpConfig.CorsConfig) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(GetCorsConfig(configCors))
}
