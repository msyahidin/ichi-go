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

func (s *ServiceImpl) PublishWelcomeNotification(ctx context.Context, userID uint32) error {
	return s.publishWelcomeNotification(ctx, userID)
}

// publishWelcomeNotification is internal implementation.
func (s *ServiceImpl) publishWelcomeNotification(ctx context.Context, userID uint32) error {
	// Check if queue system is enabled
	if s.producer == nil {
		logger.Debugf("Queue producer not configured - skipping welcome notification for user %d", userID)
		return nil
	}

	// Get user details
	user, err := s.GetById(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user for notification: %w", err)
	}

	// Build notification message
	message := userDto.WelcomeNotificationMessage{
		EventType: "user.welcome",
		UserID:    fmt.Sprintf("%d", user.ID),
		Email:     user.Email,
		Text:      fmt.Sprintf("Welcome %s! Thanks for joining us.", user.Name),
	}

	// Publish to queue
	publishOpts := rabbitmq.PublishOptions{}
	if err := s.producer.Publish(ctx, "user.welcome", message, publishOpts); err != nil {
		// Log error but don't fail the operation
		logger.Errorf("Failed to publish welcome notification for user %d: %v", userID, err)
		return fmt.Errorf("failed to publish notification: %w", err)
	}

	logger.Infof("✅ Published welcome notification for user %d", userID)
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
