package repository

import (
	"context"
	"errors"
	"ichi-go/internal/infra/database/ent"
	"ichi-go/internal/infra/database/ent/user"
	"ichi-go/pkg/logger"
	"time"
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

func (r *UserRepositoryImpl) CreateTx(ctx context.Context, txClient *ent.Client, newUser ent.User) (*ent.User, error) {
	result, err := txClient.User.Create().
		SetName(newUser.Name).
		SetEmail(newUser.Email).
		SetPassword(newUser.Password).
		SetCreatedBy(newUser.CreatedBy).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *UserRepositoryImpl) UpdateTx(ctx context.Context, txClient *ent.Client, updateUser *ent.User) (*ent.User, error) {

	affected, err := txClient.User.Update().
		Where(user.ID(updateUser.ID), user.Versions(updateUser.Versions)).
		SetName(updateUser.Name).
		SetEmail(updateUser.Email).
		SetPassword(updateUser.Password).
		SetUpdatedAt(time.Now()).
		SetCreatedBy(updateUser.CreatedBy).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	if affected < 1 {
		logger.Errorf("ID %s no records were updated in database", updateUser.ID)
		return nil, errors.New("no records were updated in database")
	}

	updated, err := txClient.User.Get(ctx, updateUser.ID)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id uint64) (*ent.User, error) {
	err := r.dbC.User.DeleteOneID(id).Exec(ctx)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *UserRepositoryImpl) SoftDelete(ctx context.Context, id uint64) (*ent.User, error) {
	deleted, err := r.dbC.User.
		UpdateOneID(id).
		// TODO : set deleted by
		//SetDeletedBy().
		SetDeletedAt(time.Now()).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return deleted, nil
}

func (r *UserRepositoryImpl) GetAll(ctx context.Context) ([]*ent.User, error) {
	data, err := r.dbC.User.Query().
		Where(user.DeletedAtIsNil()).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return data, nil
}
