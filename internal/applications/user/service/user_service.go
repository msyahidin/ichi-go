package user

import (
	"context"
	"ichi-go/db/model"
)

type Service interface {
	GetById(ctx context.Context, id uint32) (*model.User, error)
	Create(ctx context.Context, newUser model.User) (int64, error)
	Update(ctx context.Context, updateUser model.User) (int64, error)
	SendNotification(ctx context.Context, user model.User) error
}
