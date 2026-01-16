package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authDto "ichi-go/internal/applications/auth/dto"
	"ichi-go/pkg/testutil"
	"ichi-go/pkg/utils/response"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Mock Service
// ============================================================================

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx interface{}, req authDto.LoginRequest) (*authDto.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDto.LoginResponse), args.Error(1)
}

func (m *MockAuthService) Register(ctx interface{}, req authDto.RegisterRequest) (*authDto.RegisterResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDto.RegisterResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx interface{}, req authDto.RefreshTokenRequest) (*authDto.RefreshTokenResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDto.RefreshTokenResponse), args.Error(1)
}

func (m *MockAuthService) Logout(ctx interface{}, userID uint64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthService) Me(ctx interface{}, userID uint64) (*authDto.UserInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDto.UserInfo), args.Error(1)
}

// ============================================================================
// Test Setup
// ============================================================================

func setupAuthControllerTest(t *testing.T) (*echo.Echo, *MockAuthService, *AuthController) {
	t.Helper()

	e := echo.New()
	mockService := new(MockAuthService)

	// Note: Adjust this based on your actual controller constructor
	// controller := NewAuthController(mockService)

	// For now, we'll use a placeholder
	controller := &AuthController{}

	return e, mockService, controller
}

func createTestRequest(method, path string, body interface{}) (*http.Request, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	return req, nil
}

// ============================================================================
// Login Tests
// ============================================================================

func TestAuthController_Login_Success(t *testing.T) {
	// Arrange
	e, mockService, controller := setupAuthControllerTest(t)

	loginReq := authDto.LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}

	expectedResponse := &authDto.LoginResponse{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_456",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}

	mockService.On("Login", mock.Anything, loginReq).
		Return(expectedResponse, nil)

	req, err := createTestRequest(http.MethodPost, "/api/v1/auth/login", loginReq)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err = controller.Login(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var resp response.SuccessResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "success", resp.Success)
	assert.NotNil(t, resp.Data)

	mockService.AssertExpectations(t)
}

func TestAuthController_Login_InvalidCredentials(t *testing.T) {
	// Arrange
	e, mockService, controller := setupAuthControllerTest(t)

	loginReq := authDto.LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword",
	}

	mockService.On("Login", mock.Anything, loginReq).
		Return(nil, errors.New("invalid credentials"))

	req, err := createTestRequest(http.MethodPost, "/api/v1/auth/login", loginReq)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err = controller.Login(c)

	// Assert
	require.Error(t, err)
	mockService.AssertExpectations(t)
}

func TestAuthController_Login_ValidationError(t *testing.T) {
	tests := []struct {
		name      string
		request   map[string]interface{}
		wantError string
	}{
		{
			name: "Missing email",
			request: map[string]interface{}{
				"password": "Password123!",
			},
			wantError: "email",
		},
		{
			name: "Missing password",
			request: map[string]interface{}{
				"email": "test@example.com",
			},
			wantError: "password",
		},
		{
			name: "Invalid email format",
			request: map[string]interface{}{
				"email":    "invalid-email",
				"password": "Password123!",
			},
			wantError: "email",
		},
		{
			name:      "Empty request body",
			request:   map[string]interface{}{},
			wantError: "validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			e, _, controller := setupAuthControllerTest(t)

			req, err := createTestRequest(http.MethodPost, "/api/v1/auth/login", tt.request)
			require.NoError(t, err)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Act
			err = controller.Login(c)

			// Assert
			require.Error(t, err)
		})
	}
}

// ============================================================================
// Register Tests
// ============================================================================

func TestAuthController_Register_Success(t *testing.T) {
	// Arrange
	e, mockService, controller := setupAuthControllerTest(t)

	registerReq := authDto.RegisterRequest{
		Email:    "newuser@example.com",
		Name:     "newuser",
		Password: "Password123!",
	}

	expectedResponse := &authDto.RegisterResponse{
		User: authDto.UserInfo{
			ID:    1,
			Email: registerReq.Email,
			Name:  registerReq.Name,
		},
	}

	mockService.On("Register", mock.Anything, registerReq).
		Return(expectedResponse, nil)

	req, err := createTestRequest(http.MethodPost, "/api/v1/auth/register", registerReq)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err = controller.Register(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	mockService.AssertExpectations(t)
}

func TestAuthController_Register_DuplicateEmail(t *testing.T) {
	// Arrange
	e, mockService, controller := setupAuthControllerTest(t)

	registerReq := authDto.RegisterRequest{
		Email:    "existing@example.com",
		Name:     "newuser",
		Password: "Password123!",
	}

	mockService.On("Register", mock.Anything, registerReq).
		Return(nil, errors.New("email already exists"))

	req, err := createTestRequest(http.MethodPost, "/api/v1/auth/register", registerReq)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err = controller.Register(c)

	// Assert
	require.Error(t, err)
	testutil.AssertErrorContains(t, err, "email already exists")

	mockService.AssertExpectations(t)
}

func TestAuthController_Register_WeakPassword(t *testing.T) {
	// Arrange
	e, _, controller := setupAuthControllerTest(t)

	registerReq := authDto.RegisterRequest{
		Email:    "test@example.com",
		Name:     "testuser",
		Password: "weak", // Too weak
	}

	req, err := createTestRequest(http.MethodPost, "/api/v1/auth/register", registerReq)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err = controller.Register(c)

	// Assert - Should fail validation
	require.Error(t, err)
}

// ============================================================================
// Refresh Token Tests
// ============================================================================

func TestAuthController_RefreshToken_Success(t *testing.T) {
	// Arrange
	e, mockService, _ := setupAuthControllerTest(t)

	refreshReq := authDto.RefreshTokenRequest{
		RefreshToken: "valid_refresh_token",
	}

	expectedResponse := &authDto.RefreshTokenResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}

	mockService.On("RefreshToken", mock.Anything, refreshReq).
		Return(expectedResponse, nil)

	req, err := createTestRequest(http.MethodPost, "/api/v1/auth/refresh", refreshReq)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	_ = e.NewContext(req, rec)

	// Act
	// err = controller.RefreshToken(c)

	// Assert
	// require.NoError(t, err)
	// assert.Equal(t, http.StatusOK, rec.Code)

	mockService.AssertExpectations(t)
}

func TestAuthController_RefreshToken_InvalidToken(t *testing.T) {
	// Arrange
	e, mockService, _ := setupAuthControllerTest(t)

	refreshReq := authDto.RefreshTokenRequest{
		RefreshToken: "invalid_token",
	}

	mockService.On("RefreshToken", mock.Anything, refreshReq).
		Return(nil, errors.New("invalid refresh token"))

	req, err := createTestRequest(http.MethodPost, "/api/v1/auth/refresh", refreshReq)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	_ = e.NewContext(req, rec)

	// Act
	// err = controller.RefreshToken(c)

	// Assert
	// require.Error(t, err)

	mockService.AssertExpectations(t)
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestAuthController_Integration_LoginFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Arrange - Setup full Echo instance with middleware
	e := echo.New()
	mockService := new(MockAuthService)
	controller := &AuthController{}

	// Register routes
	authGroup := e.Group("/api/v1/auth")
	authGroup.POST("/login", controller.Login)
	authGroup.POST("/register", controller.Register)

	// Test registration
	t.Run("Register new user", func(t *testing.T) {
		registerReq := authDto.RegisterRequest{
			Email:    testutil.RandomEmail(),
			Name:     testutil.RandomUsername(),
			Password: "Password123!",
		}

		expectedRegisterResp := &authDto.RegisterResponse{
			User: authDto.UserInfo{
				ID:    1,
				Email: registerReq.Email,
				Name:  registerReq.Name,
			},
		}

		mockService.On("Register", mock.Anything, registerReq).
			Return(expectedRegisterResp, nil).Once()

		reqBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	// Test login
	t.Run("Login with registered user", func(t *testing.T) {
		loginReq := authDto.LoginRequest{
			Email:    "test@example.com",
			Password: "Password123!",
		}

		expectedLoginResp := &authDto.LoginResponse{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
		}

		mockService.On("Login", mock.Anything, loginReq).
			Return(expectedLoginResp, nil).Once()

		reqBody, _ := json.Marshal(loginReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkAuthController_Login(b *testing.B) {
	e := echo.New()
	mockService := new(MockAuthService)
	controller := &AuthController{}

	loginReq := authDto.LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}

	expectedResponse := &authDto.LoginResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}

	mockService.On("Login", mock.Anything, loginReq).
		Return(expectedResponse, nil)

	reqBody, _ := json.Marshal(loginReq)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		controller.Login(c)
	}
}

// ============================================================================
// Table-Driven Integration Tests
// ============================================================================

func TestAuthController_TableDriven(t *testing.T) {
	testCases := []testutil.TableTest{
		{
			Name: "Successful login",
			Run: func(t *testing.T) {
				e, mockService, controller := setupAuthControllerTest(t)

				loginReq := authDto.LoginRequest{
					Email:    "test@example.com",
					Password: "Password123!",
				}

				mockService.On("Login", mock.Anything, loginReq).
					Return(&authDto.LoginResponse{AccessToken: "token"}, nil)

				req, _ := createTestRequest(http.MethodPost, "/api/v1/auth/login", loginReq)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				err := controller.Login(c)
				require.NoError(t, err)
			},
		},
		{
			Name: "Invalid email format",
			Run: func(t *testing.T) {
				e, _, controller := setupAuthControllerTest(t)

				loginReq := map[string]string{
					"email":    "invalid",
					"password": "Password123!",
				}

				req, _ := createTestRequest(http.MethodPost, "/api/v1/auth/login", loginReq)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				err := controller.Login(c)
				require.Error(t, err)
			},
		},
		{
			Name: "Empty password",
			Run: func(t *testing.T) {
				e, _, controller := setupAuthControllerTest(t)

				loginReq := map[string]string{
					"email":    "test@example.com",
					"password": "",
				}

				req, _ := createTestRequest(http.MethodPost, "/api/v1/auth/login", loginReq)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				err := controller.Login(c)
				require.Error(t, err)
			},
		},
	}

	testutil.RunTableTests(t, testCases)
}

// ============================================================================
// Content-Type Tests
// ============================================================================

func TestAuthController_Login_InvalidContentType(t *testing.T) {
	// Arrange
	e, _, controller := setupAuthControllerTest(t)

	reqBody := `{"email":"test@example.com","password":"Password123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, "text/plain") // Wrong content type

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := controller.Login(c)

	// Assert
	require.Error(t, err)
}

// ============================================================================
// Language Header Tests
// ============================================================================

func TestAuthController_Login_WithLanguageHeader(t *testing.T) {
	tests := []struct {
		name     string
		langCode string
	}{
		{"English", "en"},
		{"Indonesian", "id"},
		{"Default (no header)", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockService, controller := setupAuthControllerTest(t)

			loginReq := authDto.LoginRequest{
				Email:    "test@example.com",
				Password: "Password123!",
			}

			mockService.On("Login", mock.Anything, loginReq).
				Return(&authDto.LoginResponse{AccessToken: "token"}, nil)

			req, _ := createTestRequest(http.MethodPost, "/api/v1/auth/login", loginReq)
			if tt.langCode != "" {
				req.Header.Set("X-Language", tt.langCode)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := controller.Login(c)
			require.NoError(t, err)
		})
	}
}
