package service

import (
	"context"
	"ichi-go/internal/applications/user/repository"
	"ichi-go/internal/infra/database/ent"
)

type UserService interface {
	GetById(ctx context.Context, id uint32) (*ent.User, error)
	BunGetById(ctx context.Context, id uint32) (*repository.UserModel, error)
}
