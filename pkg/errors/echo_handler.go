package errors

import (
	"errors"
	"github.com/labstack/echo/v4"
)

type EchoHandler struct {
}

func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (h *EchoHandler) Handle(err error, c echo.Context) error {
	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {

		// addition code for body or header here

		return nil
	}
	return err
}
