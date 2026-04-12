package middlewares

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"ichi-go/pkg/http"
	"time"
)

func AppRequestTimeOut(configHttp *http.Config) echo.MiddlewareFunc {
	return middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
		Timeout: time.Duration(configHttp.Timeout) * time.Second,
	})
}
