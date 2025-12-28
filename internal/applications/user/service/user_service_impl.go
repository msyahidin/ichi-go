package user

import (
	"context"
	"fmt"
	userDto "ichi-go/internal/applications/user/dto"
	"ichi-go/internal/applications/user/repository"
	"ichi-go/internal/infra/cache"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/clients/pokemonapi"
	"ichi-go/pkg/clients/pokemonapi/dto"
	"ichi-go/pkg/logger"
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

	// Publish welcome notification to queue
	if err := s.EnqueueWelcomeNotification(ctx, uint32(userId)); err != nil {
		// Log but don't fail user creation
		logger.Errorf("Failed to queue welcome notification: %v", err)
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

// PublishWelcomeNotification Producer publishes message to queue
func (s *ServiceImpl) PublishWelcomeNotification(ctx context.Context, userID uint32) error {
	if s.producer == nil {
		logger.Debugf("Queue not configured - skipping notification")
		return nil
	}

	user, err := s.GetById(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	message := userDto.WelcomeNotificationMessage{
		EventType: "user.welcome",
		UserID:    fmt.Sprintf("%d", user.ID),
		Email:     user.Email,
		Text:      fmt.Sprintf("Welcome %s!", user.Name),
	}

	opts := rabbitmq.PublishOptions{}
	if err := s.producer.Publish(ctx, "user.welcome", message, opts); err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	logger.Infof("Published welcome notification for user %d", userID)
	return nil
}

// Alternative naming for different use cases:

// EnqueueWelcomeNotification - Use when emphasizing it's a background job
func (s *ServiceImpl) EnqueueWelcomeNotification(ctx context.Context, userID uint32) error {
	// Same implementation as PublishWelcomeNotification
	return s.PublishWelcomeNotification(ctx, userID)
}

// SendWelcomeNotificationAsync - Use when emphasizing it's non-blocking
func (s *ServiceImpl) SendWelcomeNotificationAsync(ctx context.Context, userID uint32) error {
	// Same implementation as PublishWelcomeNotification
	return s.PublishWelcomeNotification(ctx, userID)
}
