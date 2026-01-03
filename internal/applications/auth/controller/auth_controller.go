package auth

import (
	"fmt"
	authDto "ichi-go/internal/applications/auth/dto"
	authService "ichi-go/internal/applications/auth/service"
	"ichi-go/pkg/authenticator"
	"ichi-go/pkg/logger"
	"ichi-go/pkg/utils/response"
	appValidator "ichi-go/pkg/validator"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AuthController struct {
	service *authService.ServiceImpl
}

func NewAuthController(service *authService.ServiceImpl) *AuthController {
	return &AuthController{service: service}
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user with email and password, returns access and refresh tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        X-Language  header    string                      false  "Language preference (en or id)"  default(en)
// @Param        request     body      authDto.LoginRequest        true   "Login credentials"
// @Success      200         {object}  response.SuccessResponse{data=authDto.LoginResponse}  "Login successful"
// @Failure      400         {object}  response.ErrorResponse     "Invalid request or validation error"
// @Failure      401         {object}  response.ErrorResponse     "Invalid credentials"
// @Failure      500         {object}  response.ErrorResponse     "Internal server error"
// @Router       /auth/login [post]
func (c *AuthController) Login(eCtx echo.Context) error {
	var req authDto.LoginRequest

	// Bind and validate with automatic language detection
	if err := appValidator.BindAndValidate(eCtx, &req); err != nil {
		logger.Errorf("Login request validation failed: %v", err)
		return err // Already formatted as JSON
	}

	// Authenticate user
	loginResponse, err := c.service.Login(eCtx.Request().Context(), req)
	if err != nil {
		logger.Errorf("Login failed: %v", err)
		return err
		//return response.Error(eCtx, http.StatusUnauthorized, err)
	}
	return response.Success(eCtx, loginResponse)
}

// Register godoc
// @Summary      Register new user
// @Description  Create a new user account with email, password, and name
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        X-Language  header    string                         false  "Language preference (en or id)"  default(en)
// @Param        request     body      authDto.RegisterRequest        true   "Registration data"
// @Success      201         {object}  response.SuccessResponse{data=authDto.RegisterResponse}  "User registered successfully"
// @Failure      400         {object}  response.ErrorResponse         "Invalid request, validation error, or user already exists"
// @Failure      500         {object}  response.ErrorResponse         "Internal server error"
// @Router       /auth/register [post]
func (c *AuthController) Register(eCtx echo.Context) error {
	var req authDto.RegisterRequest

	err := appValidator.BindAndValidate(eCtx, &req)
	if err != nil {
		return err
		//return response.Error(eCtx, http.StatusBadRequest, err)
	}

	registerResponse, err := c.service.Register(eCtx.Request().Context(), req)
	if err != nil {
		return err
	}

	return response.Created(eCtx, registerResponse)
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Generate new access and refresh tokens using a valid refresh token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        X-Language  header    string                              false  "Language preference (en or id)"  default(en)
// @Param        request     body      authDto.RefreshTokenRequest         true   "Refresh token"
// @Success      200         {object}  response.SuccessResponse{data=authDto.RefreshTokenResponse}  "Tokens refreshed successfully"
// @Failure      400         {object}  response.ErrorResponse              "Invalid request or validation error"
// @Failure      401         {object}  response.ErrorResponse              "Invalid or expired refresh token"
// @Failure      500         {object}  response.ErrorResponse              "Internal server error"
// @Router       /auth/refresh [post]
func (c *AuthController) RefreshToken(eCtx echo.Context) error {
	var req authDto.RefreshTokenRequest

	// Bind and validate with automatic language detection
	if err := appValidator.BindAndValidate(eCtx, &req); err != nil {
		logger.Errorf("Refresh token request validation failed: %v", err)
		return err
	}

	// Refresh tokens
	refreshResponse, err := c.service.RefreshToken(eCtx.Request().Context(), req)
	if err != nil {
		logger.Errorf("Token refresh failed: %v", err)
		return err
	}

	logger.Infof("Token refreshed successfully")
	return response.Success(eCtx, refreshResponse)
}

// Logout godoc
// @Summary      User logout
// @Description  Invalidate current user session (client should discard tokens)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.SuccessResponse{data=map[string]string}  "Logged out successfully"
// @Failure      401  {object}  response.ErrorResponse                             "Unauthorized - invalid or missing token"
// @Router       /auth/logout [post]
func (c *AuthController) Logout(eCtx echo.Context) error {
	logger.Infof("User logged out")
	return response.Success(eCtx, map[string]string{
		"message": "Logged out successfully",
	})
}

// Me godoc
// @Summary      Get current user profile
// @Description  Retrieve authenticated user's profile information
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.SuccessResponse{data=authDto.UserInfo}  "User profile retrieved successfully"
// @Failure      401  {object}  response.ErrorResponse                            "Unauthorized - invalid or missing token"
// @Failure      500  {object}  response.ErrorResponse                            "Internal server error"
// @Router       /auth/me [get]
func (c *AuthController) Me(eCtx echo.Context) error {
	fmt.Println("AuthController.Me called")
	// Get auth context from middleware
	authCtx, ok := eCtx.Get("auth").(*authenticator.AuthContext)
	logger.Debugf("AuthController.Me: authCtx=%+v, ok=%v", authCtx, ok)
	if !ok {
		return response.Error(eCtx, http.StatusUnauthorized,
			echo.NewHTTPError(http.StatusUnauthorized, "unauthorized"))
	}

	// Get user info
	user, err := c.service.Me(eCtx.Request().Context(), *authCtx)
	if err != nil {
		logger.Errorf("Failed to retrieve user profile: %v", err)
		return err
	}

	return response.Success(eCtx, user)
}
