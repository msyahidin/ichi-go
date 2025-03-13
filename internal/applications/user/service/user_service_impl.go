package service

import (
	"context"
	"ichi-go/internal/applications/user/repository"
	"ichi-go/internal/infra/database/ent"
	"ichi-go/pkg/logger"
)

type UserServiceImpl struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{repo: repo}
}

func (s *UserServiceImpl) GetById(ctx context.Context, id uint32) (*ent.User, error) {
	logger.PrintContextln(ctx, "GetById")
	return s.repo.GetById(ctx, uint64(id))
}
