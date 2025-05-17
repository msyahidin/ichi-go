package error_handler

import "github.com/labstack/echo/v4"

type ErrorHandlers []ErrorHandler

func (h ErrorHandlers) Handle(err error, ctx echo.Context) error {

	for _, handler := range h {

		if handler.Handle(err, ctx) == nil {
			return nil
		}
	}

	return err
}

func (h ErrorHandlers) EchoHandler(err error, ctx echo.Context) {

	err = h.Handle(err, ctx)
	// delegate to default handler?
}
