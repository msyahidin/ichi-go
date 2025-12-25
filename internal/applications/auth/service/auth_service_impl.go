package auth

import (
	"context"
	"errors"
	"fmt"
	authDto "ichi-go/internal/applications/auth/dto"
	userRepo "ichi-go/internal/applications/user/repository"
	"ichi-go/pkg/authenticator"

	"golang.org/x/crypto/bcrypt"
)

// Login authenticates user and returns tokens
func (s *ServiceImpl) Login(ctx context.Context, req authDto.LoginRequest) (*authDto.LoginResponse, error) {
	// Get user by email
	user, err := s.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if !s.VerifyPassword(user.Password, req.Password) {
		return nil, errors.New("invalid credentials")
	}

	// Generate tokens
	tokenPair, err := s.jwtAuth.GenerateTokens(uint64(user.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Build response
	response := &authDto.LoginResponse{
		User: authDto.UserInfo{
			ID:        uint64(user.ID),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}

	return response, nil
}

// Register creates new user and returns tokens
func (s *ServiceImpl) Register(ctx context.Context, req authDto.RegisterRequest) (*authDto.RegisterResponse, error) {
	// Check if user already exists
	existingUser, _ := s.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user model
	newUser := userRepo.UserModel{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	// Save user to database
	userID, err := s.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Get created user
	user, err := s.userRepo.GetById(ctx, uint64(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created user: %w", err)
	}

	// Generate tokens
	tokenPair, err := s.jwtAuth.GenerateTokens(uint64(user.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Build response
	response := &authDto.RegisterResponse{
		User: authDto.UserInfo{
			ID:        uint64(user.ID),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}

	return response, nil
}

// RefreshToken validates refresh token and generates new token pair
func (s *ServiceImpl) RefreshToken(ctx context.Context, req authDto.RefreshTokenRequest) (*authDto.RefreshTokenResponse, error) {
	// Validate refresh token and get user ID
	userID, err := s.jwtAuth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Verify user still exists
	user, err := s.userRepo.GetById(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// Generate new token pair
	tokenPair, err := s.jwtAuth.GenerateTokens(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Build response
	response := &authDto.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}

	return response, nil
}

// GetUserByEmail retrieves user by email address
func (s *ServiceImpl) GetUserByEmail(ctx context.Context, email string) (*userRepo.UserModel, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

// HashPassword hashes a plain text password
func (s *ServiceImpl) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword checks if provided password matches hashed password
func (s *ServiceImpl) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// Me retrieves user info by user ID
func (s *ServiceImpl) Me(ctx context.Context, authCtx authenticator.AuthContext) (*authDto.UserInfo, error) {
	user, err := s.userRepo.GetById(ctx, authCtx.UserID.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	userInfo := &authDto.UserInfo{
		ID:        uint64(user.ID),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}

	return userInfo, nil
}
