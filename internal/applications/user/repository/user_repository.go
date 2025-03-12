package repository

import (
	"context"
	"ichi-go/internal/infra/database/ent"
)

type UserRepository interface {
	GetById(ctx context.Context, id uint64) (*ent.User, error)
}
