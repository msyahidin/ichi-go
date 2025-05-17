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
}
