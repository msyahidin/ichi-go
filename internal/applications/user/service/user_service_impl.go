package service

import (
	"context"
	"rathalos-kit/internal/applications/user/repository"
)

type UserServiceImpl struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{repo: repo}
}

func (s *UserServiceImpl) GetById(ctx context.Context, id uint32) (string, error) {
	return s.repo.GetById(ctx, id)
}
