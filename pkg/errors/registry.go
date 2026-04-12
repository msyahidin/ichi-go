package errors

import (
	"ichi-go/pkg/logger"

	"github.com/labstack/echo/v5"
)

//func Setup(e *echo.Echo) {
//	e.HTTPErrorHandler = customErrorHandler
//}

func Setup(e *echo.Echo) {
	logger.Debugf("Error Handler: setting up error handler")
	e.HTTPErrorHandler = NewChain(
		NewOppsHandler(),
		NewEchoHandler(),
		NewBunHandler(),
		NewGenericHandler(),
	).EchoHandler
}
