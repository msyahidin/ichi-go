package user

import (
	"context"
	"ichi-go/db/model"
	"ichi-go/internal/infra/database/bun"
	pkgErrors "ichi-go/pkg/errors"
	"ichi-go/pkg/logger"

	upbun "github.com/uptrace/bun"
)

type RepositoryImpl struct {
	*bun.BaseRepository[model.User]
}

func NewUserRepository(dbConnection *upbun.DB) *RepositoryImpl {
	return &RepositoryImpl{BaseRepository: bun.NewRepository[model.User](dbConnection, &model.User{})}
}

func (r *RepositoryImpl) GetById(ctx context.Context, id uint64) (*model.User, error) {
	data, err := r.Find(ctx, int64(id))
	if err != nil {
		return nil, pkgErrors.Database(pkgErrors.ErrCodeDatabase).
			With("operation", "get_user").
			With("user_id", id).
			Wrap(err)
	}
	return data, nil
}

func (r *RepositoryImpl) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	data, err := r.FindBy(ctx, "email", email)
	if err != nil {
		return nil, pkgErrors.Database(pkgErrors.ErrCodeDatabase).
			With("operation", "get_user").
			With("user_email", email).
			Code(404).
			Wrap(err)
	}
	return data, nil
}

func (r *RepositoryImpl) Create(ctx context.Context, newUser model.User) (int64, error) {
	data, err := r.DB().NewInsert().
		Model(&newUser).
		Returning("id").
		Exec(ctx)
	if err != nil {
		return 0, pkgErrors.Database(pkgErrors.ErrCodeDatabase).
			With("operation", "create_user").
			With("email", newUser.Email).
			Wrap(err)
	}
	_, err = data.RowsAffected()
	if err != nil {
		return 0, pkgErrors.Database(pkgErrors.ErrCodeDatabase).
			With("operation", "create_user").
			With("email", newUser.Email).
			Wrap(err)
	}
	newUserID, err := data.LastInsertId()
	return newUserID, nil
}

func (r *RepositoryImpl) Update(ctx context.Context, updateUser model.User) (int64, error) {
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
