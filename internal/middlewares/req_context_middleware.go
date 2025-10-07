package middlewares

import (
	"github.com/labstack/echo/v4"
	"ichi-go/pkg/requestctx"
)

type ContextKey string

const (
	ContextKeyUserID   ContextKey = "user_id"
	ContextKeyIsGuest  ContextKey = "is_guest"
	ContextKeyAuthType ContextKey = "auth_type"
	ContextKeyHeaders  ContextKey = "headers"
)

func RequestContextMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rc := requestctx.FromRequest(c.Request())
			c.SetRequest(c.Request().WithContext(requestctx.NewContext(c.Request().Context(), rc)))
			return next(c)
		}
	}
}
