package repository

import "context"

type UserRepositoryImpl struct {
}

func NewUserRepository() *UserRepositoryImpl {
	return &UserRepositoryImpl{}
}

func (u *UserRepositoryImpl) GetById(ctx context.Context, id uint32) (string, error) {
	return "user", nil
}
