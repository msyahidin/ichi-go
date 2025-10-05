package middlewares

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
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
			req := c.Request()
			headers := map[string]string{}
			for k, v := range req.Header {
				if len(v) > 0 {
					headers[k] = v[0]
				}
			}

			isGuest := true
			var userID string
			var authType string

			authHeader := req.Header.Get("Authorization")
			if authHeader != "" {
				if strings.HasPrefix(authHeader, "Bearer ") {
					token := strings.TrimPrefix(authHeader, "Bearer ")

					userID = extractUserID(token)
					authType = "bearer"
					isGuest = userID == ""
				} else {
					authType = "unknown"
				}
			}

			ctx := context.WithValue(req.Context(), ContextKeyHeaders, headers)
			ctx = context.WithValue(ctx, ContextKeyUserID, userID)
			ctx = context.WithValue(ctx, ContextKeyIsGuest, isGuest)
			ctx = context.WithValue(ctx, ContextKeyAuthType, authType)

			// Replace context
			c.SetRequest(req.WithContext(ctx))

			return next(c)
		}
	}
}

func extractUserID(token string) string {
	if token == "guest" {
		return ""
	}
	return "12345"
}
