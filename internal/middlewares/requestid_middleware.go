package middlewares

import (
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
