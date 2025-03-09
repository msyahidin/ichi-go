//go:build wireinject
// +build wireinject

package user

import (
	"github.com/google/wire"
	"rathalos-kit/internal/applications/user/repository"
	"rathalos-kit/internal/applications/user/service"
)

var UserSet = wire.NewSet(
	repository.NewUserRepository,
	service.NewUserService,
	wire.Bind(new(repository.UserRepository), new(*repository.UserRepositoryImpl)),
	wire.Bind(new(service.UserService), new(*service.UserServiceImpl)),
)

func InitializedService() *service.UserServiceImpl {
	wire.Build(UserSet)
	return nil
}
