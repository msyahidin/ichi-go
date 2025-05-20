package errorhandler

import (
	"errors"
	"github.com/labstack/echo/v4"
)

type Echo struct {
}

func NewEcho() *Echo {
	return &Echo{}
}

func (h *Echo) Handle(err error, ctx echo.Context) error {

	var http *echo.HTTPError
	if errors.As(err, &http) {
		ctx.Response().Status = 422
		// addition code for body or header here

		return nil
	}

	return err
}
