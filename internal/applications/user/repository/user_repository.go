package repository

import (
	"context"
)

type UserRepository interface {
	GetById(ctx context.Context, id uint64) (*UserModel, error)
	Create(ctx context.Context, newUser UserModel) (int64, error)
	Update(ctx context.Context, updateUser UserModel) (int64, error)

	//CreateTx(ctx context.Context, txClient *ent.Client, newUser ent.User) (*ent.User, error)
	//UpdateTx(ctx context.Context, txClient *ent.Client, updateUser *ent.User) (*ent.User, error)
	//
	//Delete(ctx context.Context, id uint64) (*ent.User, error)
	//SoftDelete(ctx context.Context, id uint64) (*ent.User, error)
	//GetAll(ctx context.Context) ([]*ent.User, error)
}
