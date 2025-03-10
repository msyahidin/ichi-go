package controller

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"rathalos-kit/internal/applications/user/service"
	"rathalos-kit/pkg/utils/response"
	"strconv"
)

type UserController struct {
	service *service.UserServiceImpl
}

func NewUserController(service *service.UserServiceImpl) *UserController {
	return &UserController{service: service}
}

func (c *UserController) GetUser(eCtx echo.Context) error {
	idString, err := strconv.Atoi(eCtx.Param("id"))
	user, err := c.service.GetById(eCtx.Request().Context(), uint32(idString))
	if err != nil {
		return response.Error(eCtx, http.StatusBadRequest, err)
	}
	eCtx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return response.Success(eCtx, user)
}

func (c *UserController) GetUserPage(eCtx echo.Context) error {
	return eCtx.HTML(http.StatusOK, "<h1>This is User Page</h1>")
}
