package user

import (
	"context"
	"ichi-go/db/model"
)

type Repository interface {
	GetById(ctx context.Context, id uint64) (*model.User, error)
	Create(ctx context.Context, newUser model.User) (int64, error)
	Update(ctx context.Context, updateUser model.User) (int64, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}
