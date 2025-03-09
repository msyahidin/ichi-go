package repository

import "context"

type UserRepository interface {
	GetById(ctx context.Context, id uint32) (string, error)
}
