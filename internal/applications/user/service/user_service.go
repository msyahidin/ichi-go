package service

import (
	"context"
	"ichi-go/internal/applications/user/repository"
)

type UserService interface {
	GetById(ctx context.Context, id uint32) (*repository.UserModel, error)
	Create(ctx context.Context, newUser repository.UserModel) (int64, error)
	Update(ctx context.Context, updateUser repository.UserModel) (int64, error)
}
