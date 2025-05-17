package error_handler

import "github.com/labstack/echo/v4"

func SetupErrorHandler(e *echo.Echo) {

	e.HTTPErrorHandler = ErrorHandlers{
		NewEcho(),
		NewEnt(),
		NewGeneric(),
	}.EchoHandler
}
