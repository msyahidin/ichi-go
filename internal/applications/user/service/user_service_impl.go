package service

import (
	"context"
	"ichi-go/internal/applications/user/repository"
	"ichi-go/internal/infra/database/ent"
)

type UserServiceImpl struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{repo: repo}
}

func (s *UserServiceImpl) GetById(ctx context.Context, id uint32) (*ent.User, error) {
	return s.repo.GetById(ctx, uint64(id))
}
