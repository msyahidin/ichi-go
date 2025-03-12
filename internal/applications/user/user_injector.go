//go:build wireinject
// +build wireinject

package user

import (
	"github.com/google/wire"
	"ichi-go/internal/applications/user/repository"
	"ichi-go/internal/applications/user/service"
	"ichi-go/internal/infra/database/ent"
)

var UserSet = wire.NewSet(
	repository.NewUserRepository,
	service.NewUserService,
	wire.Bind(new(repository.UserRepository), new(*repository.UserRepositoryImpl)),
	wire.Bind(new(service.UserService), new(*service.UserServiceImpl)),
)

func InitializedService(dbConnection *ent.Client) *service.UserServiceImpl {
	wire.Build(UserSet)
	return nil
}
