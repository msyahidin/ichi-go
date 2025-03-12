package controller

import (
	dtoMapper "github.com/dranikpg/dto-mapper"
	"github.com/labstack/echo/v4"
	"ichi-go/internal/applications/user/dto"
	"ichi-go/internal/applications/user/service"
	"ichi-go/pkg/utils/response"
	"net/http"
	"strconv"
)

type UserController struct {
	service *service.UserServiceImpl
}

func NewUserController(service *service.UserServiceImpl) *UserController {
	return &UserController{service: service}
}

func (c *UserController) GetUser(eCtx echo.Context) error {
	var userGetReq dto.UserGetRequest
	err := eCtx.Bind(&userGetReq)
	if err != nil {
		return response.Error(eCtx, http.StatusBadRequest, err)
	}

	idString, err := strconv.Atoi(userGetReq.ID)
	user, err := c.service.GetById(eCtx.Request().Context(), uint32(idString))
	if err != nil {
		return response.Error(eCtx, http.StatusBadRequest, err)
	}

	var userGetResponseDto dto.UserGetResponse
	err = dtoMapper.Map(&userGetResponseDto, user)
	if err != nil {
		return response.Error(eCtx, http.StatusInternalServerError, err)
	}

	return response.Success(eCtx, userGetResponseDto)
}

func (c *UserController) GetUserPage(eCtx echo.Context) error {
	return eCtx.HTML(http.StatusOK, "<h1>This is User Page</h1>")
}
