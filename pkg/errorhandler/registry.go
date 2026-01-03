package errorhandler

import (
	"errors"
	"fmt"
	"github.com/samber/oops"
	pkgErrors "ichi-go/pkg/errors"
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

	if c.Response().Committed {
		return
	}

	code := http.StatusInternalServerError
	message := err.Error()
	fmt.Println("Custom Error Handler Invoked:", message) // this line not logged
	// Handle Echo HTTP errors
	var he *echo.HTTPError
	if errors.As(err, &he) {
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
	if oopsErr, ok := oops.AsOops(err); ok {
		code = mapErrorCodeToHTTP(oopsErr.Code())
		logger.GetInstance().Logger.
			Error().Stack().Err(err).Msg(err.Error())

	} else {
		// Non-oops errors
		logger.Errorf("Error: %v", err)
	}

	response.Error(c, code, err)
}

func mapErrorCodeToHTTP(code interface{}) int {
	codeStr := fmt.Sprintf("%v", code)

	switch codeStr {
	case pkgErrors.ErrCodeUserExists:
		return http.StatusConflict // 409
	case pkgErrors.ErrCodeInvalidCredentials:
		return http.StatusUnauthorized // 401
	case pkgErrors.ErrCodeUserNotFound:
		return http.StatusNotFound // 404
	case pkgErrors.ErrCodeInvalidToken:
		return http.StatusUnauthorized // 401
	case pkgErrors.ErrCodeValidation:
		return http.StatusBadRequest // 400
	case pkgErrors.ErrCodeDatabase:
		return http.StatusInternalServerError // 500
	default:
		return http.StatusInternalServerError // 500
	}
}
