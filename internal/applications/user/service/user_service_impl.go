package user

import (
	"context"
	"fmt"
	user "ichi-go/internal/applications/user/repository"
	"ichi-go/internal/infra/cache"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/clients/pokemonapi"
	"ichi-go/pkg/clients/pokemonapi/dto"
	"ichi-go/pkg/db/model"
	"time"
)

type ServiceImpl struct {
	repo       user.Repository
	cache      cache.Cache
	pokeClient pokemonapi.PokemonClient
	producer   rabbitmq.MessageProducer
}

func NewUserService(
	repo user.Repository,
	cache cache.Cache,
	pokeClient pokemonapi.PokemonClient,
	producer rabbitmq.MessageProducer, // Renamed from msgConnection
) *ServiceImpl {
	return &ServiceImpl{
		repo:       repo,
		cache:      cache,
		pokeClient: pokeClient,
		producer:   producer,
	}
}

func (s *ServiceImpl) GetById(ctx context.Context, id uint32) (*model.User, error) {
	cacheKey := fmt.Sprintf("user:%d", id)
	cachedData, err := s.cache.Get(ctx, cacheKey, &model.User{})
	if err == nil && cachedData != nil {
		return cachedData.(*model.User), nil
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

func (s *ServiceImpl) Create(ctx context.Context, newUser model.User) (int64, error) {
	userId, err := s.repo.Create(ctx, newUser)
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (s *ServiceImpl) Update(ctx context.Context, updateUser model.User) (int64, error) {
	user, err := s.repo.GetById(ctx, uint64(updateUser.ID))
	if err != nil {
		return 0, err
	}
	if user == nil || user.ID == 0 {
		return 0, fmt.Errorf("user with ID %d not found", updateUser.ID)
	}

	updatedId, err := s.repo.Update(ctx, updateUser)
	if err != nil {
		return 0, err
	}
	cacheKey := fmt.Sprintf("user:%d", updateUser.ID)
	_, _ = s.cache.Delete(ctx, cacheKey)
	return updatedId, nil
}

func (s *ServiceImpl) GetPokemon(ctx context.Context, name string) (*dto.PokemonGetResponseDto, error) {
	return s.pokeClient.GetDetail(ctx, name)
}
