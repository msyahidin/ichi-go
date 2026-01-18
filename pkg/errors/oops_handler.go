package errors

import (
	"fmt"
	"ichi-go/pkg/logger"
	"ichi-go/pkg/utils/response"
	appValidator "ichi-go/pkg/validator"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/samber/oops"
)

type OopsErrorHandler struct {
}

func NewOppsHandler() *OopsErrorHandler {
	return &OopsErrorHandler{}
}

func (h *OopsErrorHandler) Handle(err error, c echo.Context) error {
	if c.Response().Committed {
		return nil
	}
	code := http.StatusInternalServerError

	//// Handle Echo HTTP errors
	//var he *echo.HTTPError
	//if errors.As(err, &he) {
	//	code = he.Code
	//	message := fmt.Sprintf("%v", he.Message)
	//}

	// Handle validation errors
	if validationErr := appValidator.GetValidationError(err); validationErr != nil {
		code = http.StatusBadRequest
		logger.Warnf("Validation error: %v", err)
		response.Error(c, code, err)
		return err
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
	return err
}

func mapErrorCodeToHTTP(code interface{}) int {
	codeStr := fmt.Sprintf("%v", code)

	switch codeStr {
	// Auth - 401 Unauthorized
	case ErrCodeInvalidCredentials,
		ErrCodeInvalidToken,
		ErrCodeTokenExpired,
		ErrCodeUnauthorized:
		return http.StatusUnauthorized

	// Auth - 400 Bad Request
	case ErrCodePasswordWeak:
		return http.StatusBadRequest

	// Auth - 404 Not Found
	case ErrCodeUserNotFound,
		ErrCodeNotFound:
		return http.StatusNotFound

	// Auth - 409 Conflict
	case ErrCodeUserExists:
		return http.StatusConflict

	// Validation - 400 Bad Request
	case ErrCodeValidation:
		return http.StatusBadRequest

	// Infrastructure - 500 Internal Server Error
	case ErrCodeDatabase,
		ErrCodeCache,
		ErrCodeQueue,
		ErrCodeInternal,
		ErrCodePasswordHashFailed,
		ErrCodeTokenGenFailed,
		ErrCodeNotificationFailed:
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}
