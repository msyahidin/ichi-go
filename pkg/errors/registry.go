package errors

import (
	"github.com/labstack/echo/v4"
	"ichi-go/pkg/logger"
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
