package errorhandler

import "github.com/labstack/echo/v4"

type ErrorHandler interface {
	Handle(err error, ctx echo.Context) error
}

type Chain []ErrorHandler

func NewChain(handlers ...ErrorHandler) Chain {
	return handlers
}

func (c Chain) Handle(err error, ctx echo.Context) error {
	for _, h := range c {
		if h.Handle(err, ctx) == nil {
			return nil
		}
	}
	return err
}

func (c Chain) EchoHandler(err error, ctx echo.Context) {
	if remaining := c.Handle(err, ctx); remaining != nil {
		NewGenericHandler().Handle(remaining, ctx)
	}
}
