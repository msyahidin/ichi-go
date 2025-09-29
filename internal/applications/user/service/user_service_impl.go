package service

import (
	"context"
	"fmt"
	"ichi-go/internal/applications/user/repository"
	"ichi-go/internal/infra/cache"
	"ichi-go/internal/infra/database/ent"
	"ichi-go/pkg/clients/pokemonapi"
	"ichi-go/pkg/clients/pokemonapi/dto"
	"time"
)

type UserServiceImpl struct {
	repo       repository.UserRepository
	cache      cache.Cache
	pokeClient pokemonapi.PokemonClient
}

func NewUserService(repo repository.UserRepository, cache cache.Cache) *UserServiceImpl {
	pokeClient := pokemonapi.NewPokemonClientImpl()
	return &UserServiceImpl{repo: repo, cache: cache, pokeClient: pokeClient}
}

func (s *UserServiceImpl) GetById(ctx context.Context, id uint32) (*ent.User, error) {
	user := ent.User{
		ID:        0,
		Versions:  0,
		Name:      "",
		Email:     "",
		Password:  "",
		CreatedAt: time.Time{},
		CreatedBy: 0,
		UpdatedAt: time.Time{},
		UpdatedBy: 0,
		DeletedAt: time.Time{},
		DeletedBy: 0,
	}
	//cacheKey := fmt.Sprintf("user:%d", id)
	//cachedData, err := s.cache.Get(ctx, cacheKey, &ent.User{})
	//if err == nil && cachedData != nil {
	//	return cachedData.(*ent.User), nil
	//}
	//
	//user, err := s.repo.GetById(ctx, uint64(id))
	//if err != nil {
	//	return nil, err
	//}
	//
	//if err == nil {
	//	option := cache.Options{
	//		Expiration: time.Hour,
	//		Compress:   true,
	//	}
	//	_, _ = s.cache.Set(ctx, cacheKey, user, option)
	//}
	return &user, nil
}

func (s *UserServiceImpl) BunGetById(ctx context.Context, id uint32) (*repository.UserModel, error) {

	cacheKey := fmt.Sprintf("user:%d", id)
	cachedData, err := s.cache.Get(ctx, cacheKey, &repository.UserModel{})
	if err == nil && cachedData != nil {
		return cachedData.(*repository.UserModel), nil
	}

	user, err := s.repo.GetById(ctx, uint64(id))
	if err != nil {
		return nil, err
	}

	if err == nil {
		option := cache.Options{
			Expiration: time.Hour,
			Compress:   true,
		}
		_, _ = s.cache.Set(ctx, cacheKey, user, option)
	}
	return user, nil
}

func (s *UserServiceImpl) GetPokemon(ctx context.Context, name string) (*dto.PokemonGetResponseDto, error) {
	return s.pokeClient.GetDetail(ctx, name)
}
