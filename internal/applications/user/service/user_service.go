package service

import (
	"context"
	"rathalos-kit/internal/infrastructure/database/ent"
)

type UserService interface {
	GetById(ctx context.Context, id uint32) (*ent.User, error)
}
