package user

import (
	"context"
	"fmt"
	"ichi-go/internal/applications/user/repository"
	"ichi-go/internal/infra/cache"
	"ichi-go/internal/infra/messaging/rabbitmq"
	"ichi-go/pkg/clients/pokemonapi"
	"ichi-go/pkg/clients/pokemonapi/dto"
	"time"
)

type ServiceImpl struct {
	repo       user.Repository
	cache      cache.Cache
	pokeClient pokemonapi.PokemonClient
	mc         *rabbitmq.Connection
}

func NewUserService(repo user.Repository, cache cache.Cache, pokeClient pokemonapi.PokemonClient, mc *rabbitmq.Connection) *ServiceImpl {
	return &ServiceImpl{repo: repo, cache: cache, pokeClient: pokeClient, mc: mc}
}

func (s *ServiceImpl) GetById(ctx context.Context, id uint32) (*user.UserModel, error) {
	cacheKey := fmt.Sprintf("user:%d", id)
	cachedData, err := s.cache.Get(ctx, cacheKey, &user.UserModel{})
	if err == nil && cachedData != nil {
		return cachedData.(*user.UserModel), nil
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

func (s *ServiceImpl) Create(ctx context.Context, newUser user.UserModel) (int64, error) {
	userId, err := s.repo.Create(ctx, newUser)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func (s *ServiceImpl) Update(ctx context.Context, updateUser user.UserModel) (int64, error) {
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

func (s *ServiceImpl) SendNotification(ctx context.Context, userID uint32) error {

	//user, err := s.GetById(ctx, userID)
	//
	//if err != nil {
	//	return errors.New("user not found")
	//}
	//
	//if s.mc == nil {
	//	return errors.New("messaging connection is not initialized")
	//}
	//
	//cfg := config.Get()
	//
	//publisher, err := rabbitmq.NewPublisher(s.mc, cfg.Messaging().RabbitMQ)
	//
	//if err != nil {
	//	return err
	//}
	//msg := dtoUser.WelcomeNotificationMessage{
	//	EventType: "user.welcome",
	//	UserId:    fmt.Sprintf("%d", user.ID),
	//	Email:     user.Email,
	//	Text:      fmt.Sprintf("Welcome %s to Ichi-Go!", user.Name),
	//}
	//var publishOpt = rabbitmq.PublishOptions{}
	//err = publisher.Publish(ctx, "user.welcome", msg, publishOpt)
	//if err != nil {
	//	return err
	//}
	return nil
}
