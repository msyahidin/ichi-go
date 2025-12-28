package user

import (
	"context"
	upbun "github.com/uptrace/bun"
	"ichi-go/internal/infra/database/bun"
	"ichi-go/pkg/logger"
)

type RepositoryImpl struct {
	*bun.BaseRepository[UserModel]
}

func NewUserRepository(dbConnection *upbun.DB) *RepositoryImpl {
	return &RepositoryImpl{BaseRepository: bun.NewRepository[UserModel](dbConnection, &UserModel{})}
}

func (r *RepositoryImpl) GetById(ctx context.Context, id uint64) (*UserModel, error) {
	data, err := r.Find(ctx, int64(id))
	if err != nil {
		logger.Errorf("Error user repo with data: %+v, err: %+v", data, err)
		return nil, err
	}
	return data, nil
}

func (r *RepositoryImpl) FindByEmail(ctx context.Context, email string) (*UserModel, error) {
	data, err := r.FindBy(ctx, "email", email)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *RepositoryImpl) Create(ctx context.Context, newUser UserModel) (int64, error) {
	data, err := r.DB().NewInsert().
		Model(&newUser).
		Returning("id").
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
	return newUserID, nil
}

func (r *RepositoryImpl) Update(ctx context.Context, updateUser UserModel) (int64, error) {
	data, err := r.DB().NewUpdate().
		Model(&updateUser).
		Where("id = ?", updateUser.ID).
		OmitZero().
		Returning("id").
		Exec(ctx)
	if err != nil {
		logger.Errorf("Error user repo with data: %+v, err: %+v", updateUser, err)
		return 0, err
	}
	rowsAffected, err := data.RowsAffected()
	if err != nil {
		logger.Errorf("Error getting rows affected when updating user with data: %+v, err: %+v", updateUser, err)
		return 0, err
	}
	if rowsAffected < 1 {
		logger.Errorf("ID %d no records were updated in database", updateUser.ID)
		return 0, nil
	}
	logger.Debugf("User updated with result: %+v", data)
	return updateUser.ID, nil
}
