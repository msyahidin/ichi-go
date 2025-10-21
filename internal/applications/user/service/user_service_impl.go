package service

import (
	"context"
	"errors"
	"fmt"
	"ichi-go/config"
	dtoUser "ichi-go/internal/applications/user/dto"
	"ichi-go/internal/applications/user/repository"
	"ichi-go/internal/infra/cache"
	"ichi-go/internal/infra/messaging/rabbitmq"
	"ichi-go/pkg/clients/pokemonapi"
	"ichi-go/pkg/clients/pokemonapi/dto"
	"time"
)

type UserServiceImpl struct {
	repo       repository.UserRepository
	cache      cache.Cache
	pokeClient pokemonapi.PokemonClient
	mc         *rabbitmq.Connection
}

func NewUserService(repo repository.UserRepository, cache cache.Cache, mc *rabbitmq.Connection) *UserServiceImpl {
	pokeClient := pokemonapi.NewPokemonClientImpl()
	return &UserServiceImpl{repo: repo, cache: cache, pokeClient: pokeClient, mc: mc}
}

func (s *UserServiceImpl) GetById(ctx context.Context, id uint32) (*repository.UserModel, error) {

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

func (s *UserServiceImpl) Create(ctx context.Context, newUser repository.UserModel) (int64, error) {
	userId, err := s.repo.Create(ctx, newUser)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func (s *UserServiceImpl) Update(ctx context.Context, updateUser repository.UserModel) (int64, error) {
	user, err := s.repo.GetById(ctx, uint64(updateUser.ID))
	if err != nil {
		return 0, err
	}
	if user == nil || user.ID == 0 {
		return 0, fmt.Errorf("user with ID %d not found", updateUser.ID)
	}

	updatedId, err := s.repo.Update(ctx, *user)
	if err != nil {
		return 0, err
	}
	cacheKey := fmt.Sprintf("user:%d", updateUser.ID)
	_, _ = s.cache.Delete(ctx, cacheKey)
	return updatedId, nil
}

func (s *UserServiceImpl) GetPokemon(ctx context.Context, name string) (*dto.PokemonGetResponseDto, error) {
	return s.pokeClient.GetDetail(ctx, name)
}

func (s *UserServiceImpl) SendNotification(ctx context.Context, id uint32) error {
	userModel, err := s.repo.GetById(ctx, uint64(id))
	if err != nil {
		return errors.New("failed to get user")
	}
	if s.mc == nil {
		return errors.New("messaging connection is not initialized")
	}

	cfg := config.Get()

	publisher, err := rabbitmq.NewPublisher(s.mc, cfg.Messaging().RabbitMQ)

	if err != nil {
		return err
	}
	msg := dtoUser.WelcomeNotificationMessage{
		EventType: "user.welcome",
		UserId:    fmt.Sprintf("%d", userModel.ID),
		Email:     userModel.Email,
		Text:      fmt.Sprintf("Welcome %s to Ichi-Go!", userModel.Name),
	}

	err = publisher.Publish(ctx, "user.welcome", msg)
	if err != nil {
		return err
	}
	return nil
}
