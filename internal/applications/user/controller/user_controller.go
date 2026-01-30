package user

import (
	userDto "ichi-go/internal/applications/user/dto"
	userService "ichi-go/internal/applications/user/service"
	pokeDto "ichi-go/pkg/clients/pokemonapi/dto"
	"ichi-go/pkg/db/model"
	"ichi-go/pkg/logger"
	"ichi-go/pkg/requestctx"
	mapper "ichi-go/pkg/utils/dto"
	"ichi-go/pkg/utils/response"
	"net/http"
	"strconv"

	dtoMapper "github.com/dranikpg/dto-mapper"
	"github.com/labstack/echo/v4"
)

type UserController struct {
	service *userService.ServiceImpl
}

func NewUserController(service *userService.ServiceImpl) *UserController {
	return &UserController{service: service}
}

func (c *UserController) GetUser(eCtx echo.Context) error {
	var userGetReq userDto.UserGetRequest
	logger.Debugf("GetUser request: %+v", requestctx.FromContext(eCtx.Request().Context()))
	err := eCtx.Bind(&userGetReq)
	if err != nil {
		return response.Error(eCtx, http.StatusBadRequest, err)
	}

	idString, err := strconv.Atoi(userGetReq.ID)
	user, err := c.service.GetById(eCtx.Request().Context(), uint32(idString))
	if err != nil {
		return response.Error(eCtx, http.StatusBadRequest, err)
	}
	var userGetResponseDto userDto.UserGetResponse
	dtoMap := mapper.New()
	err = dtoMap.SafeMap(&userGetResponseDto, user)
	if err != nil {
		return response.Error(eCtx, http.StatusInternalServerError, err)
	}

	return response.Success(eCtx, user)
}

func (c *UserController) CreateUser(eCtx echo.Context) error {
	var userCreateReq userDto.UserCreateRequest
	err := eCtx.Bind(&userCreateReq)
	if err != nil {
		return response.Error(eCtx, http.StatusBadRequest, err)
	}

	var userModel model.User
	err = dtoMapper.Map(&userModel, userCreateReq)
	if err != nil {
		return response.Error(eCtx, http.StatusInternalServerError, err)
	}
	user, err := c.service.Create(eCtx.Request().Context(), userModel)
	if err != nil {
		return response.Error(eCtx, http.StatusInternalServerError, err)
	}

	return response.Created(eCtx, map[string]interface{}{"new_user_id": user})
}

func (c *UserController) UpdateUser(eCtx echo.Context) error {
	var userUpdateReq userDto.UserUpdateRequest
	err := eCtx.Bind(&userUpdateReq)
	if err != nil {
		return response.Error(eCtx, http.StatusBadRequest, err)
	}

	var userModel model.User
	err = dtoMapper.Map(&userModel, userUpdateReq)
	if err != nil {
		return response.Error(eCtx, http.StatusInternalServerError, err)
	}
	user, err := c.service.Update(eCtx.Request().Context(), userModel)
	if err != nil {
		return response.Error(eCtx, http.StatusInternalServerError, err)
	}

	return response.Success(eCtx, user)
}

func (c *UserController) GetUserPage(eCtx echo.Context) error {
	return eCtx.HTML(http.StatusOK, "<h1>This is User Page</h1>")
}

func (c *UserController) GetPokemon(eCtx echo.Context) error {
	name := eCtx.Param("name")
	var pokemonGetResponseDto pokeDto.PokemonGetResponseDto
	result, err := c.service.GetPokemon(eCtx.Request().Context(), name)
	if err != nil {
		return response.Error(eCtx, http.StatusInternalServerError, err)
	}
	err = dtoMapper.Map(&pokemonGetResponseDto, result)
	return response.Success(eCtx, pokemonGetResponseDto)
}
