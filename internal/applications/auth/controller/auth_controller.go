package auth

import (
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

// Login handles user login
// POST /api/auth/login
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
		return response.Error(eCtx, http.StatusUnauthorized, err)
	}

	logger.Infof("User logged in successfully: %s", req.Email)
	return response.Success(eCtx, loginResponse)
}

// Register handles user registration
// POST /api/auth/register
func (c *AuthController) Register(eCtx echo.Context) error {
	var req authDto.RegisterRequest

	// Bind and validate with automatic language detection
	err := appValidator.BindAndValidate(eCtx, &req)
	logger.Errorf("Register request validation result: %v", err)
	if err != nil {
		logger.Errorf("Register request validation failed: %v", err)
		return response.Error(eCtx, http.StatusBadRequest, err) // Already formatted as JSON
	}
	return err

	//logger.Errorf("Register request validation passed %v", err)

	//// Register user
	//registerResponse, err := c.service.Register(eCtx.Request().Context(), req)
	//if err != nil {
	//	logger.Errorf("Registration failed: %v", err)
	//	return response.Error(eCtx, http.StatusBadRequest, err)
	//}
	//
	//logger.Infof("User registered successfully: %s", req.Email)
	//return response.Created(eCtx, "success")
}

// RefreshToken handles token refresh
// POST /api/auth/refresh
func (c *AuthController) RefreshToken(eCtx echo.Context) error {
	var req authDto.RefreshTokenRequest

	// Bind and validate with automatic language detection
	if err := appValidator.BindAndValidate(eCtx, &req); err != nil {
		logger.Errorf("Refresh token request validation failed: %v", err)
		return err // Already formatted as JSON
	}

	// Refresh tokens
	refreshResponse, err := c.service.RefreshToken(eCtx.Request().Context(), req)
	if err != nil {
		logger.Errorf("Token refresh failed: %v", err)
		return response.Error(eCtx, http.StatusUnauthorized, err)
	}

	logger.Infof("Token refreshed successfully")
	return response.Success(eCtx, refreshResponse)
}

// Logout handles user logout
// POST /api/auth/logout
func (c *AuthController) Logout(eCtx echo.Context) error {
	logger.Infof("User logged out")
	return response.Success(eCtx, map[string]string{
		"message": "Logged out successfully",
	})
}

// Me returns current authenticated user info
// GET /api/auth/me
func (c *AuthController) Me(eCtx echo.Context) error {
	// Get auth context from middleware
	authCtx, ok := eCtx.Get("auth").(*authenticator.AuthContext)
	if !ok {
		return response.Error(eCtx, http.StatusUnauthorized,
			echo.NewHTTPError(http.StatusUnauthorized, "unauthorized"))
	}

	// Get user info
	user, err := c.service.Me(eCtx.Request().Context(), *authCtx)
	if err != nil {
		logger.Errorf("Failed to get user info: %v", err)
		return response.Error(eCtx, http.StatusInternalServerError, err)
	}

	return response.Success(eCtx, user)
}
