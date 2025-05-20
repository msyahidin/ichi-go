package error_handler

import (
	"github.com/labstack/echo/v4"
	"ichi-go/internal/infra/database/ent"
)

type Ent struct {
}

func NewEnt() *Ent {
	return &Ent{}
}

func (h *Ent) Handle(err error, ctx echo.Context) error {

	if ent.IsNotFound(err) {
		ctx.Response().Status = 404
		// addition code for body or header here
		return nil
	}

	if ent.IsConstraintError(err) {
		ctx.Response().Status = 500
		// addition code for body or header here
		return nil
	}

	return err
package errorhandler

import (
    "github.com/labstack/echo/v4"
    "ichi-go/internal/infra/database/ent"
)

type EntHandler struct {
}

func NewEntHandler() *EntHandler {
    return &EntHandler{}
}

// Handle maps Ent errors to HTTP responses; otherwise it returns the error
// for the next handler in the chain.
func (h *EntHandler) Handle(err error, c echo.Context) error {
    switch {
    case ent.IsNotFound(err):
       ctx.Response().Status = 404
		// addition code for body or header here
		return nil

    case ent.IsConstraintError(err):
        ctx.Response().Status = 500
		// addition code for body or header here
		return nil
    }

    // not an Ent error: let next handler deal with it
    return err
}
