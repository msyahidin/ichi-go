package errorhandler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"ichi-go/pkg/logger"
	"ichi-go/pkg/utils/response"
	appValidator "ichi-go/pkg/validator"
)

func Setup(e *echo.Echo) {
	e.HTTPErrorHandler = customErrorHandler
}

func customErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := err.Error()

	// Handle Echo HTTP errors
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = fmt.Sprintf("%v", he.Message)
	}

	// Handle validation errors
	if validationErr := appValidator.GetValidationError(err); validationErr != nil {
		code = http.StatusBadRequest
		logger.Warnf("Validation error: %v", err)
		response.Error(c, code, err)
		return
	}

	// Handle oops errors
	//if oopsErr, ok := oops.AsOops(err); ok {
	//	logger.Error().
	//		Interface("code", oopsErr.Code()).
	//		Interface("context", oopsErr.Context()).
	//		Strs("tags", oopsErr.Tags()).
	//		Str("hint", oopsErr.Hint()).
	//		Str("trace", oopsErr.Trace()).
	//		Err(err).
	//		Msg(message)
	//} else {
	logger.Errorf("Error: %v. Message: %v", err, message)
	//}

	response.Error(c, code, err)
}
