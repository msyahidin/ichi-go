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

	cacheKey := fmt.Sprintf("user:%d", id)
	cachedData, err := s.cache.Get(ctx, cacheKey, &ent.User{})
	if err == nil && cachedData != nil {
		return cachedData.(*ent.User), nil
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
