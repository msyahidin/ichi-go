package auth

import (
	"context"
	"fmt"
	authDto "ichi-go/internal/applications/auth/dto"
	userDto "ichi-go/internal/applications/user/dto"
	userRepo "ichi-go/internal/applications/user/repository"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/authenticator"
	pkgErrors "ichi-go/pkg/errors"
	"ichi-go/pkg/logger"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Login authenticates user and returns tokens
func (s *ServiceImpl) Login(ctx context.Context, req authDto.LoginRequest) (*authDto.LoginResponse, error) {
	user, err := s.GetUserByEmail(ctx, req.Email)
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
			With("user_id", user.ID).
			Hint("Invalid email or password").
			Errorf("password verification failed")
	}

	tokenPair, err := s.jwtAuth.GenerateTokens(uint64(user.ID))
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeInternal).
			With("user_id", user.ID).
			Wrap(err)
	}

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
	existingUser, _ := s.GetUserByEmail(ctx, req.Email)
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

	tokenPair, err := s.jwtAuth.GenerateTokens(uint64(user.ID))
	if err != nil {
		return nil, pkgErrors.AuthService(pkgErrors.ErrCodeInternal).
			With("user_id", user.ID).
			Wrap(err)
	}

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

	if err := s.EnqueueWelcomeNotification(ctx, uint32(userID)); err != nil {
		logger.Errorf("%v", pkgErrors.Queue("NOTIFICATION_FAILED").
			With("user_id", userID).
			Wrap(err))
	}

	return response, nil
}

// RefreshToken validates refresh token and generates new token pair
func (s *ServiceImpl) RefreshToken(ctx context.Context, req authDto.RefreshTokenRequest) (*authDto.RefreshTokenResponse, error) {
	userID, err := s.jwtAuth.ValidateRefreshToken(req.RefreshToken)
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

	tokenPair, err := s.jwtAuth.GenerateTokens(userID)
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
		ID:        uint64(user.ID),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

// EnqueueWelcomeNotification PublishWelcomeNotification Producer publishes message to queue
func (s *ServiceImpl) EnqueueWelcomeNotification(ctx context.Context, userID uint32) error {
	if s.producer == nil {
		logger.Debugf("Queue not configured - skipping notification")
		return nil
	}

	user, err := s.userRepo.GetById(ctx, uint64(userID))
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	message := userDto.WelcomeNotificationMessage{
		EventType: "user.welcome",
		UserID:    fmt.Sprintf("%d", user.ID),
		Email:     user.Email,
		Text:      fmt.Sprintf("Welcome %s!", user.Name),
	}

	opts := rabbitmq.PublishOptions{
		Delay: 30 * time.Second,
	}
	if err := s.producer.Publish(ctx, "user.welcome", message, opts); err != nil {
		logger.Errorf("Failed to publish welcome notification: %v", err)
		return fmt.Errorf("failed to publish: %w", err)
	}
	return nil
}
