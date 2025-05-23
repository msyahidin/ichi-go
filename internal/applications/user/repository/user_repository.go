package repository

import (
	"context"
	"ichi-go/internal/infra/database/ent"
)

type UserRepository interface {
	GetById(ctx context.Context, id uint64) (*ent.User, error)

	CreateTx(ctx context.Context, txClient *ent.Client, newUser ent.User) (*ent.User, error)
	UpdateTx(ctx context.Context, txClient *ent.Client, updateUser *ent.User) (*ent.User, error)

	Delete(ctx context.Context, id uint64) (*ent.User, error)
	SoftDelete(ctx context.Context, id uint64) (*ent.User, error)
	GetAll(ctx context.Context) ([]*ent.User, error)
}
