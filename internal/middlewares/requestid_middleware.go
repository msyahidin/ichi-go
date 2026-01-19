package middlewares

import (
	"context"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/random"
)

func AppRequestID() echo.MiddlewareFunc {

	return middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: RequestIDGenerator,
	})
}

func RequestIDGenerator() string {
	requestID, err := uuid.NewRandom()
	if err != nil {
		return random.String(32)
	}
	return requestID.String()
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
