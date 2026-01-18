package errors

import (
	"database/sql"
	"errors"
	"ichi-go/pkg/logger"

	"github.com/labstack/echo/v4"
)

type BunErrorHandler struct {
}

func NewBunHandler() *BunErrorHandler {
	return &BunErrorHandler{}
}

// Handle maps Bun errors to HTTP responses; otherwise it returns the error
func (h *BunErrorHandler) Handle(err error, ctx echo.Context) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		logger.Debugf("BunErrorHandler: record not found")
		ctx.Response().Status = 404
		return nil

	case errors.Is(err, sql.ErrTxDone):
		ctx.Response().Status = 500
		return nil
	}
	return err
}
