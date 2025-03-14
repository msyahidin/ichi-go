package middlewares

import (
	"github.com/labstack/echo/v4"
)

type RequestFields struct {
	Domain string
}

func RequestInjector(fields RequestFields) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Request().Header.Set("Domain", fields.Domain)
			return next(c)
		}
	}
}
