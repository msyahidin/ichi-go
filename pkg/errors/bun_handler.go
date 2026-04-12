package errors

import (
	"database/sql"
	"errors"
	"ichi-go/pkg/logger"

	"github.com/labstack/echo/v5"
)

type BunErrorHandler struct {
}

func NewBunHandler() *BunErrorHandler {
	return &BunErrorHandler{}
}

// Handle maps Bun errors to HTTP responses; otherwise it returns the error
func (h *BunErrorHandler) Handle(err error, ctx *echo.Context) error {
	resp, _ := echo.UnwrapResponse(ctx.Response())
	switch {
	case errors.Is(err, sql.ErrNoRows):
		logger.Debugf("BunErrorHandler: record not found")
		if resp != nil {
			resp.Status = 404
		}
		return nil

	case errors.Is(err, sql.ErrTxDone):
		if resp != nil {
			resp.Status = 500
		}
		return nil
	}
	return err
}
