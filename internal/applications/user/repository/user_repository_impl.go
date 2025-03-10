package repository

import (
	"context"
	"errors"
	"rathalos-kit/internal/infrastructure/database/ent"
	"rathalos-kit/internal/infrastructure/database/ent/user"
)

type UserRepositoryImpl struct {
	dbC *ent.Client //non transactional client
}

func NewUserRepository(dbConnection *ent.Client) *UserRepositoryImpl {
	return &UserRepositoryImpl{dbC: dbConnection}
}

func (r *UserRepositoryImpl) GetById(ctx context.Context, id uint64) (*ent.User, error) {
	data, err := r.dbC.User.Query().
		Where(user.And(
			user.ID(id),
		)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return data, err
}
