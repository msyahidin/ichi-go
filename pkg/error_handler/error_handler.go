package error_handler

import "github.com/labstack/echo/v4"

type ErrorHandler interface {
	Handle(err error, ctx echo.Context) error
}
