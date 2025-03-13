// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package user

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"ichi-go/internal/applications/user/repository"
	"ichi-go/internal/applications/user/service"
	"ichi-go/internal/infra/cache"
	"ichi-go/internal/infra/database/ent"
)

// Injectors from user_injector.go:

func InitializedService(dbConnection *ent.Client, cacheConnection *redis.Client) *service.UserServiceImpl {
	userRepositoryImpl := repository.NewUserRepository(dbConnection)
	cacheImpl := cache.NewCache(cacheConnection)
	userServiceImpl := service.NewUserService(userRepositoryImpl, cacheImpl)
	return userServiceImpl
}

// user_injector.go:

var UserSet = wire.NewSet(repository.NewUserRepository, service.NewUserService, cache.NewCache, wire.Bind(new(repository.UserRepository), new(*repository.UserRepositoryImpl)), wire.Bind(new(service.UserService), new(*service.UserServiceImpl)), wire.Bind(new(cache.Cache), new(*cache.CacheImpl)))
