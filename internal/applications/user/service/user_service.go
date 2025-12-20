package user

import (
	"context"
	"ichi-go/internal/applications/user/repository"
)

type Service interface {
	GetById(ctx context.Context, id uint32) (*user.UserModel, error)
	Create(ctx context.Context, newUser user.UserModel) (int64, error)
	Update(ctx context.Context, updateUser user.UserModel) (int64, error)
	SendNotification(ctx context.Context, user user.UserModel) error
}
