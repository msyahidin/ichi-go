package errorhandler

import (
	"github.com/labstack/echo/v4"
)

type GenericHandler struct {
}

func NewGenericHandler() *GenericHandler {
	return &GenericHandler{}
}

func (g *GenericHandler) Handle(err error, ctx echo.Context) error {

	ctx.Response().Status = 500
	// addition code for body or header here
	return nil
}
