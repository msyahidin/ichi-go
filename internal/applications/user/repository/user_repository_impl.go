package repository

import (
	"context"
	upbun "github.com/uptrace/bun"
	"ichi-go/internal/infra/database/bun"
	"ichi-go/pkg/logger"
)

type UserRepositoryImpl struct {
	*bun.BaseRepository[UserModel]
}

func NewUserRepository(dbConnection *upbun.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{BaseRepository: bun.NewRepository[UserModel](dbConnection, &UserModel{})}
}

func (r *UserRepositoryImpl) GetById(ctx context.Context, id uint64) (*UserModel, error) {
	data, err := r.Find(ctx, int64(id))
	if err != nil {
		logger.Errorf("Error user repo with data: %+v, err: %+v", data, err)
		return nil, err
	}
	return data, nil
}

func (r *UserRepositoryImpl) Create(ctx context.Context, newUser UserModel) (int64, error) {
	data, err := r.DB().NewInsert().Model(&newUser).
		Exec(ctx)
	if err != nil {
		logger.Errorf("Error user repo with data: %+v, err: %+v", newUser, err)
		return 0, err
	}
	_, err = data.RowsAffected()
	if err != nil {
		logger.Errorf("Error getting rows affected when creating user with data: %+v, err: %+v", newUser, err)
		return 0, err
	}
	newUserID, err := data.LastInsertId()
	logger.Debugf("User created with result: %+v", data)
	return newUserID, nil
}

//
//func (r *UserRepositoryImpl) CreateTx(ctx context.Context, txClient *ent.Client, newUser ent.User) (*ent.User, error) {
//	result, err := txClient.User.Create().
//		SetName(newUser.Name).
//		SetEmail(newUser.Email).
//		SetPassword(newUser.Password).
//		SetCreatedBy(newUser.CreatedBy).
//		Save(ctx)
//
//	if err != nil {
//		return nil, err
//	}
//
//	return result, nil
//}
//
//func (r *UserRepositoryImpl) UpdateTx(ctx context.Context, txClient *ent.Client, updateUser *ent.User) (*ent.User, error) {
//
//	affected, err := txClient.User.Update().
//		Where(user.ID(updateUser.ID), user.Versions(updateUser.Versions)).
//		SetName(updateUser.Name).
//		SetEmail(updateUser.Email).
//		SetPassword(updateUser.Password).
//		SetUpdatedAt(time.Now()).
//		SetCreatedBy(updateUser.CreatedBy).
//		Save(ctx)
//
//	if err != nil {
//		return nil, err
//	}
//
//	if affected < 1 {
//		logger.Errorf("ID %d no records were updated in database", updateUser.ID)
//		return nil, errors.New("no records were updated in database")
//	}
//
//	updated, err := txClient.User.Get(ctx, updateUser.ID)
//	if err != nil {
//		return nil, err
//	}
//
//	return updated, nil
//}
//
//func (r *UserRepositoryImpl) Delete(ctx context.Context, id uint64) (*ent.User, error) {
//	err := r.dbC.User.DeleteOneID(id).Exec(ctx)
//
//	if err != nil {
//		return nil, err
//	}
//
//	return nil, nil
//}
//
//func (r *UserRepositoryImpl) SoftDelete(ctx context.Context, id uint64) (*ent.User, error) {
//	deleted, err := r.dbC.User.
//		UpdateOneID(id).
//		// TODO : set deleted by
//		//SetDeletedBy().
//		SetDeletedAt(time.Now()).
//		Save(ctx)
//
//	if err != nil {
//		return nil, err
//	}
//
//	return deleted, nil
//}
//
//func (r *UserRepositoryImpl) GetAll(ctx context.Context) ([]*ent.User, error) {
//	data, err := r.dbC.User.Query().
//		Where(user.DeletedAtIsNil()).
//		All(ctx)
//
//	if err != nil {
//		return nil, err
//	}
//
//	return data, nil
//}
