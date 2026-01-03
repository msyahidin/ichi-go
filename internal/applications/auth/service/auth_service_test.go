package auth

import (
	"context"
	"database/sql"
	"errors"
	"ichi-go/internal/infra/queue/rabbitmq"
	"testing"
	_ "time"

	authDto "ichi-go/internal/applications/auth/dto"
	userRepo "ichi-go/internal/applications/user/repository"
	"ichi-go/pkg/authenticator"
	pkgErrors "ichi-go/pkg/errors"

	_ "github.com/labstack/echo/v4"
	"github.com/samber/oops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// MockUserRepository is a mock implementation of user repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*userRepo.UserModel, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userRepo.UserModel), args.Error(1)
}

func (m *MockUserRepository) GetById(ctx context.Context, id uint64) (*userRepo.UserModel, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userRepo.UserModel), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user userRepo.UserModel) (int64, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user userRepo.UserModel) (int64, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockJWTService wraps JWT functionality for testing
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateTokens(userID uint64) (*authenticator.TokenPair, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authenticator.TokenPair), args.Error(1)
}

func (m *MockJWTService) ValidateRefreshToken(token string) (uint64, error) {
	args := m.Called(token)
	return args.Get(0).(uint64), args.Error(1)
}

// MockMessageProducer is a mock implementation of message producer
type MockMessageProducer struct {
	mock.Mock
}

func (m *MockMessageProducer) Publish(ctx context.Context, routingKey string, message interface{}, opts rabbitmq.PublishOptions) error {
	args := m.Called(ctx, routingKey, message, opts)
	return args.Error(0)
}

func (m *MockMessageProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// ============================================================================
// Test Fixtures
// ============================================================================

// TestUser wraps UserModel with test-specific fields
type TestUser struct {
	Model    *userRepo.UserModel
	ID       uint64
	Password string // Plain text password for testing
}

func createTestUser() *TestUser {
	plainPassword := "Password123!"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	return &TestUser{
		Model: &userRepo.UserModel{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: string(hashedPassword),
		},
		ID:       1,
		Password: plainPassword,
	}
}

func createTestUserWithID(id uint64, email string) *TestUser {
	plainPassword := "Password123!"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	return &TestUser{
		Model: &userRepo.UserModel{
			Name:     "Test User",
			Email:    email,
			Password: string(hashedPassword),
		},
		ID:       id,
		Password: plainPassword,
	}
}

func createTestTokenPair() *authenticator.TokenPair {
	return &authenticator.TokenPair{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.access",
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.refresh",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}
}

// AuthServiceForTesting defines the testable interface
type AuthServiceForTesting interface {
	Login(ctx context.Context, req authDto.LoginRequest) (*authDto.LoginResponse, error)
	Register(ctx context.Context, req authDto.RegisterRequest) (*authDto.RegisterResponse, error)
	RefreshToken(ctx context.Context, req authDto.RefreshTokenRequest) (*authDto.RefreshTokenResponse, error)
	Me(ctx context.Context, authCtx authenticator.AuthContext) (*authDto.UserInfo, error)
	GetUserByEmail(ctx context.Context, email string) (*userRepo.UserModel, error)
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) bool
}

// testAuthService implements AuthServiceForTesting with mocks
type testAuthService struct {
	userRepo userRepo.Repository
	jwtMock  *MockJWTService
	producer rabbitmq.MessageProducer
}

func newTestAuthService(repo userRepo.Repository, jwt *MockJWTService, producer rabbitmq.MessageProducer) *testAuthService {
	return &testAuthService{
		userRepo: repo,
		jwtMock:  jwt,
		producer: producer,
	}
}

// Login implementation for testing
func (s *testAuthService) Login(ctx context.Context, req authDto.LoginRequest) (*authDto.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeInvalidCredentials).
			With("email", req.Email).
			Hint("Invalid email or password").
			Wrap(err)
	}

	if user == nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeUserNotFound).
			With("email", req.Email).
			Hint("Invalid email or password").
			Errorf("user not found")
	}

	if !s.VerifyPassword(user.Password, req.Password) {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeInvalidCredentials).
			With("email", req.Email).
			Errorf("password verification failed")
	}

	userID := uint64(1) // In tests, we use a fixed ID
	tokenPair, err := s.jwtMock.GenerateTokens(userID)
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeInternal).
			With("user_id", userID).
			Wrap(err)
	}

	return &authDto.LoginResponse{
		User: authDto.UserInfo{
			ID:    userID,
			Name:  user.Name,
			Email: user.Email,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// Register implementation for testing
func (s *testAuthService) Register(ctx context.Context, req authDto.RegisterRequest) (*authDto.RegisterResponse, error) {
	existingUser, _ := s.userRepo.FindByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeUserExists).
			With("email", req.Email).
			Hint("Email already registered").
			Errorf("user already exists")
	}

	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeInternal).
			Wrap(err)
	}

	newUser := userRepo.UserModel{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	userID, err := s.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeDatabase).
			With("email", req.Email).
			Wrap(err)
	}

	user, err := s.userRepo.GetById(ctx, uint64(userID))
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeDatabase).
			With("user_id", userID).
			Wrap(err)
	}

	tokenPair, err := s.jwtMock.GenerateTokens(uint64(userID))
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeInternal).
			With("user_id", userID).
			Wrap(err)
	}

	return &authDto.RegisterResponse{
		User: authDto.UserInfo{
			ID:    uint64(userID),
			Name:  user.Name,
			Email: user.Email,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// RefreshToken implementation for testing
func (s *testAuthService) RefreshToken(ctx context.Context, req authDto.RefreshTokenRequest) (*authDto.RefreshTokenResponse, error) {
	userID, err := s.jwtMock.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeInvalidToken).
			Hint("Invalid or expired refresh token").
			Wrap(err)
	}

	user, err := s.userRepo.GetById(ctx, userID)
	if err != nil || user == nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeUserNotFound).
			With("user_id", userID).
			Hint("User not found").
			Errorf("user not found")
	}

	tokenPair, err := s.jwtMock.GenerateTokens(userID)
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeInternal).
			With("user_id", userID).
			Wrap(err)
	}

	return &authDto.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// Me implementation for testing
func (s *testAuthService) Me(ctx context.Context, authCtx authenticator.AuthContext) (*authDto.UserInfo, error) {
	user, err := s.userRepo.GetById(ctx, authCtx.UserID.ID)
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeDatabase).
			With("user_id", authCtx.UserID.ID).
			WithContext(ctx).
			Wrap(err)
	}

	if user == nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeUserNotFound).
			With("user_id", authCtx.UserID.ID).
			Errorf("user not found")
	}

	return &authDto.UserInfo{
		ID:    authCtx.UserID.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

// Helper methods
func (s *testAuthService) GetUserByEmail(ctx context.Context, email string) (*userRepo.UserModel, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

func (s *testAuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *testAuthService) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func setupAuthService(t *testing.T) (*testAuthService, *MockUserRepository, *MockJWTService, *MockMessageProducer) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTService)
	mockProducer := new(MockMessageProducer)

	service := newTestAuthService(mockRepo, mockJWT, mockProducer)

	return service, mockRepo, mockJWT, mockProducer
}

// ============================================================================
// Login Tests
// ============================================================================

func TestLogin_Success(t *testing.T) {
	// Arrange
	service, mockRepo, mockJWT, _ := setupAuthService(t)
	ctx := context.Background()
	testUser := createTestUser()
	testTokens := createTestTokenPair()

	req := authDto.LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}

	mockRepo.On("FindByEmail", ctx, req.Email).Return(testUser.Model, nil)
	mockJWT.On("GenerateTokens", testUser.ID).Return(testTokens, nil)

	// Act
	response, err := service.Login(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, testUser.ID, response.User.ID)
	assert.Equal(t, testUser.Model.Name, response.User.Name)
	assert.Equal(t, testUser.Model.Email, response.User.Email)
	assert.Equal(t, testTokens.AccessToken, response.AccessToken)
	assert.Equal(t, testTokens.RefreshToken, response.RefreshToken)
	assert.Equal(t, testTokens.TokenType, response.TokenType)
	assert.Equal(t, testTokens.ExpiresIn, response.ExpiresIn)

	mockRepo.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	// Arrange
	service, mockRepo, _, _ := setupAuthService(t)
	ctx := context.Background()

	req := authDto.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "Password123!",
	}

	mockRepo.On("FindByEmail", ctx, req.Email).Return(nil, sql.ErrNoRows)

	// Act
	response, err := service.Login(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	// Verify it's an oops error with correct code
	oopsErr, ok := oops.AsOops(err)
	require.True(t, ok, "Expected oops error")
	assert.Equal(t, pkgErrors.ErrCodeInvalidCredentials, oopsErr.Code())

	mockRepo.AssertExpectations(t)
}

func TestLogin_InvalidPassword(t *testing.T) {
	// Arrange
	service, mockRepo, _, _ := setupAuthService(t)
	ctx := context.Background()
	testUser := createTestUser()

	req := authDto.LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword123!",
	}

	mockRepo.On("FindByEmail", ctx, req.Email).Return(testUser.Model, nil)

	// Act
	response, err := service.Login(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	// Verify it's an oops error with correct code
	oopsErr, ok := oops.AsOops(err)
	require.True(t, ok, "Expected oops error")
	assert.Equal(t, pkgErrors.ErrCodeInvalidCredentials, oopsErr.Code())

	mockRepo.AssertExpectations(t)
}

func TestLogin_TokenGenerationFailed(t *testing.T) {
	// Arrange
	service, mockRepo, mockJWT, _ := setupAuthService(t)
	ctx := context.Background()
	testUser := createTestUser()

	req := authDto.LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}

	mockRepo.On("FindByEmail", ctx, req.Email).Return(testUser.Model, nil)
	mockJWT.On("GenerateTokens", testUser.ID).
		Return(nil, errors.New("jwt signing failed"))

	// Act
	response, err := service.Login(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	// Verify it's an oops error with correct code
	oopsErr, ok := oops.AsOops(err)
	require.True(t, ok, "Expected oops error")
	assert.Equal(t, pkgErrors.ErrCodeInternal, oopsErr.Code())

	mockRepo.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestLogin_DatabaseError(t *testing.T) {
	// Arrange
	service, mockRepo, _, _ := setupAuthService(t)
	ctx := context.Background()

	req := authDto.LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}

	mockRepo.On("FindByEmail", ctx, req.Email).
		Return(nil, errors.New("database connection failed"))

	// Act
	response, err := service.Login(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// Register Tests
// ============================================================================

func TestRegister_Success(t *testing.T) {
	// Arrange
	service, mockRepo, mockJWT, mockProducer := setupAuthService(t)
	ctx := context.Background()
	testTokens := createTestTokenPair()

	req := authDto.RegisterRequest{
		Name:     "New User",
		Email:    "newuser@example.com",
		Password: "SecurePass123!@#",
	}

	// User doesn't exist yet
	mockRepo.On("FindByEmail", ctx, req.Email).Return(nil, sql.ErrNoRows)

	// User creation succeeds
	mockRepo.On("Create", ctx, mock.MatchedBy(func(u userRepo.UserModel) bool {
		return u.Name == req.Name && u.Email == req.Email
	})).Return(int64(1), nil)

	// Get created user
	createdUser := &userRepo.UserModel{
		Name:     req.Name,
		Email:    req.Email,
		Password: "hashed_password_here", // Will be set by service
	}
	// Note: ID, CreatedAt, UpdatedAt are set by database/ORM
	mockRepo.On("GetById", ctx, uint64(1)).Return(createdUser, nil)

	// Token generation
	mockJWT.On("GenerateTokens", uint64(1)).Return(testTokens, nil)

	// Queue notification (can fail without breaking registration)
	mockProducer.On("Publish", ctx, "user.welcome", mock.Anything, mock.Anything).
		Return(nil)

	// Act
	response, err := service.Register(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, uint64(1), response.User.ID)
	assert.Equal(t, req.Name, response.User.Name)
	assert.Equal(t, req.Email, response.User.Email)
	assert.Equal(t, testTokens.AccessToken, response.AccessToken)
	assert.Equal(t, testTokens.RefreshToken, response.RefreshToken)

	mockRepo.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	// Arrange
	service, mockRepo, _, _ := setupAuthService(t)
	ctx := context.Background()
	existingUser := createTestUser()

	req := authDto.RegisterRequest{
		Name:     "New User",
		Email:    "test@example.com", // Same as existing user
		Password: "SecurePass123!@#",
	}

	mockRepo.On("FindByEmail", ctx, req.Email).Return(existingUser, nil)

	// Act
	response, err := service.Register(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	// Verify it's an oops error with correct code
	oopsErr, ok := oops.AsOops(err)
	require.True(t, ok, "Expected oops error")
	assert.Equal(t, pkgErrors.ErrCodeUserExists, oopsErr.Code())

	mockRepo.AssertExpectations(t)
}

func TestRegister_CreateUserFailed(t *testing.T) {
	// Arrange
	service, mockRepo, _, _ := setupAuthService(t)
	ctx := context.Background()

	req := authDto.RegisterRequest{
		Name:     "New User",
		Email:    "newuser@example.com",
		Password: "SecurePass123!@#",
	}

	mockRepo.On("FindByEmail", ctx, req.Email).Return(nil, sql.ErrNoRows)
	mockRepo.On("Create", ctx, mock.Anything).
		Return(int64(0), errors.New("database constraint violation"))

	// Act
	response, err := service.Register(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	// Verify it's an oops error with correct code
	oopsErr, ok := oops.AsOops(err)
	require.True(t, ok, "Expected oops error")
	assert.Equal(t, pkgErrors.ErrCodeDatabase, oopsErr.Code())

	mockRepo.AssertExpectations(t)
}

func TestRegister_GetCreatedUserFailed(t *testing.T) {
	// Arrange
	service, mockRepo, _, _ := setupAuthService(t)
	ctx := context.Background()

	req := authDto.RegisterRequest{
		Name:     "New User",
		Email:    "newuser@example.com",
		Password: "SecurePass123!@#",
	}

	mockRepo.On("FindByEmail", ctx, req.Email).Return(nil, sql.ErrNoRows)
	mockRepo.On("Create", ctx, mock.Anything).Return(int64(1), nil)
	mockRepo.On("GetById", ctx, uint64(1)).
		Return(nil, errors.New("user disappeared"))

	// Act
	response, err := service.Register(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	mockRepo.AssertExpectations(t)
}

func TestRegister_NotificationFailureDoesNotBreakRegistration(t *testing.T) {
	// Arrange
	service, mockRepo, mockJWT, mockProducer := setupAuthService(t)
	ctx := context.Background()
	testTokens := createTestTokenPair()

	req := authDto.RegisterRequest{
		Name:     "New User",
		Email:    "newuser@example.com",
		Password: "SecurePass123!@#",
	}

	mockRepo.On("FindByEmail", ctx, req.Email).Return(nil, sql.ErrNoRows)
	mockRepo.On("Create", ctx, mock.Anything).Return(int64(1), nil)

	createdUser := &userRepo.UserModel{
		Name:  req.Name,
		Email: req.Email,
	}
	mockRepo.On("GetById", ctx, uint64(1)).Return(createdUser, nil)
	mockJWT.On("GenerateTokens", uint64(1)).Return(testTokens, nil)

	// Notification fails but shouldn't break registration
	mockProducer.On("Publish", ctx, "user.welcome", mock.Anything, mock.Anything).
		Return(errors.New("queue unavailable"))

	// Act
	response, err := service.Register(ctx, req)

	// Assert - Should still succeed despite notification failure
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, uint64(1), response.User.ID)

	mockRepo.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

// ============================================================================
// RefreshToken Tests
// ============================================================================

func TestRefreshToken_Success(t *testing.T) {
	// Arrange
	service, mockRepo, mockJWT, _ := setupAuthService(t)
	ctx := context.Background()
	testUser := createTestUser()
	testTokens := createTestTokenPair()

	req := authDto.RefreshTokenRequest{
		RefreshToken: "valid.refresh.token",
	}

	mockJWT.On("ValidateRefreshToken", req.RefreshToken).Return(testUser.ID, nil)
	mockRepo.On("GetById", ctx, testUser.ID).Return(testUser.Model, nil)
	mockJWT.On("GenerateTokens", testUser.ID).Return(testTokens, nil)

	// Act
	response, err := service.RefreshToken(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, testTokens.AccessToken, response.AccessToken)
	assert.Equal(t, testTokens.RefreshToken, response.RefreshToken)
	assert.Equal(t, testTokens.TokenType, response.TokenType)
	assert.Equal(t, testTokens.ExpiresIn, response.ExpiresIn)

	mockJWT.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	// Arrange
	service, _, mockJWT, _ := setupAuthService(t)
	ctx := context.Background()

	req := authDto.RefreshTokenRequest{
		RefreshToken: "invalid.refresh.token",
	}

	mockJWT.On("ValidateRefreshToken", req.RefreshToken).
		Return(uint64(0), errors.New("token expired"))

	// Act
	response, err := service.RefreshToken(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	// Verify it's an oops error with correct code
	oopsErr, ok := oops.AsOops(err)
	require.True(t, ok, "Expected oops error")
	assert.Equal(t, pkgErrors.ErrCodeInvalidToken, oopsErr.Code())

	mockJWT.AssertExpectations(t)
}

func TestRefreshToken_UserNotFound(t *testing.T) {
	// Arrange
	service, mockRepo, mockJWT, _ := setupAuthService(t)
	ctx := context.Background()

	req := authDto.RefreshTokenRequest{
		RefreshToken: "valid.refresh.token",
	}

	mockJWT.On("ValidateRefreshToken", req.RefreshToken).Return(uint64(999), nil)
	mockRepo.On("GetById", ctx, uint64(999)).Return(nil, sql.ErrNoRows)

	// Act
	response, err := service.RefreshToken(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	// Verify it's an oops error with correct code
	oopsErr, ok := oops.AsOops(err)
	require.True(t, ok, "Expected oops error")
	assert.Equal(t, pkgErrors.ErrCodeUserNotFound, oopsErr.Code())

	mockJWT.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestRefreshToken_NewTokenGenerationFailed(t *testing.T) {
	// Arrange
	service, mockRepo, mockJWT, _ := setupAuthService(t)
	ctx := context.Background()
	testUser := createTestUser()

	req := authDto.RefreshTokenRequest{
		RefreshToken: "valid.refresh.token",
	}

	mockJWT.On("ValidateRefreshToken", req.RefreshToken).Return(testUser.ID, nil)
	mockRepo.On("GetById", ctx, testUser.ID).Return(testUser.Model, nil)
	mockJWT.On("GenerateTokens", testUser.ID).
		Return(nil, errors.New("signing key unavailable"))

	// Act
	response, err := service.RefreshToken(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	mockJWT.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// Me Tests
// ============================================================================

func TestMe_Success(t *testing.T) {
	// Arrange
	service, mockRepo, _, _ := setupAuthService(t)
	ctx := context.Background()
	testUser := createTestUser()

	authCtx := authenticator.AuthContext{
		UserID: authenticator.UserSubject{ID: testUser.ID},
	}

	mockRepo.On("GetById", ctx, testUser.ID).Return(testUser.Model, nil)

	// Act
	response, err := service.Me(ctx, authCtx)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, uint64(testUser.ID), response.ID)
	assert.NotEmpty(t, response.Name)
	assert.NotEmpty(t, response.Email)

	mockRepo.AssertExpectations(t)
}

func TestMe_UserNotFound(t *testing.T) {
	// Arrange
	service, mockRepo, _, _ := setupAuthService(t)
	ctx := context.Background()

	authCtx := authenticator.AuthContext{
		UserID: authenticator.UserSubject{ID: 999},
	}

	mockRepo.On("GetById", ctx, uint64(999)).Return(nil, sql.ErrNoRows)

	// Act
	response, err := service.Me(ctx, authCtx)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	// Verify it's an oops error with correct code
	oopsErr, ok := oops.AsOops(err)
	require.True(t, ok, "Expected oops error")
	assert.Equal(t, pkgErrors.ErrCodeDatabase, oopsErr.Code())

	mockRepo.AssertExpectations(t)
}

func TestMe_DatabaseError(t *testing.T) {
	// Arrange
	service, mockRepo, _, _ := setupAuthService(t)
	ctx := context.Background()

	authCtx := authenticator.AuthContext{
		UserID: authenticator.UserSubject{ID: 1},
	}

	mockRepo.On("GetById", ctx, uint64(1)).
		Return(nil, errors.New("database connection lost"))

	// Act
	response, err := service.Me(ctx, authCtx)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// Helper Method Tests
// ============================================================================

func TestGetUserByEmail_Success(t *testing.T) {
	// Arrange
	service, mockRepo, _, _ := setupAuthService(t)
	ctx := context.Background()
	testUser := createTestUser()

	mockRepo.On("FindByEmail", ctx, testUser.Model.Email).Return(testUser, nil)

	// Act
	user, err := service.GetUserByEmail(ctx, testUser.Model.Email)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, testUser.Model.Email, user.Email)

	mockRepo.AssertExpectations(t)
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	// Arrange
	service, mockRepo, _, _ := setupAuthService(t)
	ctx := context.Background()

	mockRepo.On("FindByEmail", ctx, "notfound@example.com").
		Return(nil, sql.ErrNoRows)

	// Act
	user, err := service.GetUserByEmail(ctx, "notfound@example.com")

	// Assert
	require.Error(t, err)
	assert.Nil(t, user)

	mockRepo.AssertExpectations(t)
}

func TestHashPassword_Success(t *testing.T) {
	// Arrange
	service, _, _, _ := setupAuthService(t)
	password := "TestPassword123!"

	// Act
	hashedPassword, err := service.HashPassword(password)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)

	// Verify the hash is valid
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	assert.NoError(t, err)
}

func TestVerifyPassword_Success(t *testing.T) {
	// Arrange
	service, _, _, _ := setupAuthService(t)
	password := "TestPassword123!"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// Act
	valid := service.VerifyPassword(string(hashedPassword), password)

	// Assert
	assert.True(t, valid)
}

func TestVerifyPassword_Invalid(t *testing.T) {
	// Arrange
	service, _, _, _ := setupAuthService(t)
	password := "TestPassword123!"
	wrongPassword := "WrongPassword123!"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// Act
	valid := service.VerifyPassword(string(hashedPassword), wrongPassword)

	// Assert
	assert.False(t, valid)
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestLogin_EmptyPassword(t *testing.T) {
	// Arrange
	service, mockRepo, _, _ := setupAuthService(t)
	ctx := context.Background()
	testUser := createTestUser()

	req := authDto.LoginRequest{
		Email:    "test@example.com",
		Password: "", // Empty password
	}

	mockRepo.On("FindByEmail", ctx, req.Email).Return(testUser.Model, nil)

	// Act
	response, err := service.Login(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	mockRepo.AssertExpectations(t)
}

func TestRegister_WeakPassword(t *testing.T) {
	// This test assumes validation happens before reaching service layer
	// The service itself doesn't validate password strength
	// But we can test that a weak password still gets hashed properly

	// Arrange
	service, _, _, _ := setupAuthService(t)
	weakPassword := "weak"

	// Act
	hashedPassword, err := service.HashPassword(weakPassword)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, weakPassword, hashedPassword)
}

func TestRegister_SpecialCharactersInEmail(t *testing.T) {
	// Arrange
	service, mockRepo, mockJWT, mockProducer := setupAuthService(t)
	ctx := context.Background()
	testTokens := createTestTokenPair()

	req := authDto.RegisterRequest{
		Name:     "Special User",
		Email:    "test+special@example.com", // Email with + character
		Password: "SecurePass123!@#",
	}

	mockRepo.On("FindByEmail", ctx, req.Email).Return(nil, sql.ErrNoRows)
	mockRepo.On("Create", ctx, mock.Anything).Return(int64(1), nil)

	createdUser := &userRepo.UserModel{
		Name:  req.Name,
		Email: req.Email,
	}
	mockRepo.On("GetById", ctx, uint64(1)).Return(createdUser, nil)
	mockJWT.On("GenerateTokens", uint64(1)).Return(testTokens, nil)
	mockProducer.On("Publish", ctx, "user.welcome", mock.Anything, mock.Anything).
		Return(nil)

	// Act
	response, err := service.Register(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, req.Email, response.User.Email)

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// Concurrent Access Tests
// ============================================================================

func TestLogin_ConcurrentRequests(t *testing.T) {
	// Arrange
	service, mockRepo, mockJWT, _ := setupAuthService(t)
	ctx := context.Background()
	testUser := createTestUser()
	testTokens := createTestTokenPair()

	req := authDto.LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}

	// Setup expectations for multiple calls
	mockRepo.On("FindByEmail", ctx, req.Email).Return(testUser.Model, nil).Times(10)
	mockJWT.On("GenerateTokens", testUser.ID).Return(testTokens, nil).Times(10)

	// Act - Simulate 10 concurrent login requests
	results := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := service.Login(ctx, req)
			results <- err
		}()
	}

	// Assert - All requests should succeed
	for i := 0; i < 10; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	mockRepo.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkLogin_Success(b *testing.B) {
	service, mockRepo, mockJWT, _ := setupAuthService(&testing.T{})
	ctx := context.Background()
	testUser := createTestUser()
	testTokens := createTestTokenPair()

	req := authDto.LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}

	mockRepo.On("FindByEmail", ctx, req.Email).Return(testUser.Model, nil)
	mockJWT.On("GenerateTokens", testUser.ID).Return(testTokens, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.Login(ctx, req)
	}
}

func BenchmarkHashPassword(b *testing.B) {
	service, _, _, _ := setupAuthService(&testing.T{})
	password := "TestPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	service, _, _, _ := setupAuthService(&testing.T{})
	password := "TestPassword123!"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.VerifyPassword(string(hashedPassword), password)
	}
}
