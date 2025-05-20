package error_handler

import (
	"github.com/labstack/echo/v4"
)

type Generic struct {
}

func NewGeneric() *Generic {
	return &Generic{}
}

func (g *Generic) Handle(err error, ctx echo.Context) error {

	ctx.Response().Status = 500
	// addition code for body or header here
	return nil
package errorhandler

import (
	"github.com/labstack/echo/v4"
)

type GenericHandler struct {
}

func NewGeneric() *GenericHandler {
	return &GenericHandler{}
}

func (g *Generic) Handle(err error, ctx echo.Context) error {

	ctx.Response().Status = 500
	// addition code for body or header here
	return nil
}
