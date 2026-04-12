package errors

import (
	"github.com/labstack/echo/v5"
)

type GenericHandler struct {
}

func NewGenericHandler() *GenericHandler {
	return &GenericHandler{}
}

func (g *GenericHandler) Handle(err error, ctx *echo.Context) error {

	if resp, unwrapErr := echo.UnwrapResponse(ctx.Response()); unwrapErr == nil {
		resp.Status = 500
	}
	// addition code for body or header here
	return nil
}
