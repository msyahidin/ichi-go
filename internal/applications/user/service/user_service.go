package service

import "context"

type UserService interface {
	GetById(ctx context.Context, id uint32) (string, error)
}
