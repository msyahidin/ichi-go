package middlewares

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func AppRequestID() echo.MiddlewareFunc {

	return middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: RequestIDGenerator,
	})
}

func RequestIDGenerator() string {
	requestID, err := uuid.NewRandom()
	if err != nil {
		return fmt.Sprintf("%016x%016x", rand.Int63(), rand.Int63())
	}
	return requestID.String()
}

func copyRequestID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		requestID := c.Request().Header.Get(echo.HeaderXRequestID)
		if requestID == "" {
			requestID = c.Response().Header().Get(echo.HeaderXRequestID)
		}
		ctx := context.WithValue(c.Request().Context(), echo.HeaderXRequestID, requestID)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
