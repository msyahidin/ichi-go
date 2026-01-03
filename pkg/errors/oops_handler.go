package errors

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/samber/oops"
	"ichi-go/pkg/logger"
	"ichi-go/pkg/utils/response"
	appValidator "ichi-go/pkg/validator"
	"net/http"
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
	case ErrCodeUserExists:
		return http.StatusConflict // 409
	case ErrCodeInvalidCredentials:
		return http.StatusUnauthorized // 401
	case ErrCodeUserNotFound:
		return http.StatusNotFound // 404
	case ErrCodeInvalidToken:
		return http.StatusUnauthorized // 401
	case ErrCodeValidation:
		return http.StatusBadRequest // 400
	case ErrCodeDatabase:
		return http.StatusInternalServerError // 500
	default:
		return http.StatusInternalServerError // 500
	}
}
