package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"ichi-go/config"
	"net/http"
)

func GetCorsConfig() middleware.CORSConfig {
	allowedOrigins := config.Cfg.Http.Cors.AllowOrigins
	return middleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
	}
}

func Cors() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(GetCorsConfig())
}
