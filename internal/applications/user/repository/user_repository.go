package repository

import (
	"context"
	"rathalos-kit/internal/infrastructure/database/ent"
)

type UserRepository interface {
	GetById(ctx context.Context, id uint64) (*ent.User, error)
}
