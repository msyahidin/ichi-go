package errorhandler

import "github.com/labstack/echo/v4"

func Setup(e *echo.Echo) {
	e.HTTPErrorHandler = NewChain(
		NewEchoHandler(),
		NewEntHandler(),     // dari ent_handler.go
		NewGenericHandler(), // dari generic_handler.go
	).EchoHandler
}
