package user

import (
	"context"
)

type Repository interface {
	GetById(ctx context.Context, id uint64) (*UserModel, error)
	Create(ctx context.Context, newUser UserModel) (int64, error)
	Update(ctx context.Context, updateUser UserModel) (int64, error)
	FindByEmail(ctx context.Context, email string) (*UserModel, error)
}
