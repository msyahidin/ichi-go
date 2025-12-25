package auth

import (
	"context"
	authDto "ichi-go/internal/applications/auth/dto"
	userRepo "ichi-go/internal/applications/user/repository"
	"ichi-go/pkg/authenticator"
)

type Service interface {
	Login(ctx context.Context, req authDto.LoginRequest) (*authDto.LoginResponse, error)
	Register(ctx context.Context, req authDto.RegisterRequest) (*authDto.RegisterResponse, error)
	RefreshToken(ctx context.Context, req authDto.RefreshTokenRequest) (*authDto.RefreshTokenResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*userRepo.UserModel, error)
	VerifyPassword(hashedPassword, password string) bool
	Me(ctx context.Context, userId uint64) (*authDto.UserInfo, error)
}

type ServiceImpl struct {
	userRepo userRepo.Repository
	jwtAuth  *authenticator.JWTAuthenticator
}

func NewAuthService(userRepo userRepo.Repository, jwtAuth *authenticator.JWTAuthenticator) *ServiceImpl {
	return &ServiceImpl{
		userRepo: userRepo,
		jwtAuth:  jwtAuth,
	}
}
