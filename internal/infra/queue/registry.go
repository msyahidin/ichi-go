package queue

import (
	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	"github.com/spf13/viper"
	"github.com/uptrace/bun"

	notifChannels "ichi-go/internal/applications/notification/channels"
	notifConsumers "ichi-go/internal/applications/notification/consumers"
	"ichi-go/internal/applications/notification/repositories"
	"ichi-go/internal/applications/notification/services"
	orderConsumers "ichi-go/internal/applications/order/consumers"
	userConsumers "ichi-go/internal/applications/user/consumers"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/logger"
	"ichi-go/pkg/notification/fcm"
	notiftemplate "ichi-go/pkg/notification/template"
)

// ConsumerRegistration links consumer name to processing function.
type ConsumerRegistration struct {
	Name        string               // Must match config.yaml
	ConsumeFunc rabbitmq.ConsumeFunc // Processing function
	Description string               // What this consumer does
}

// GetRegisteredConsumers returns all queue consumers.
// The injector is required to resolve dependencies for consumers that need DB or services.
//
// To add new consumer:
// 1. Create in internal/applications/{domain}/consumers/
// 2. Implement Consume(ctx, body) error
// 3. Add registration here
// 4. Add config in config.yaml
// 5. Test
func GetRegisteredConsumers(injector do.Injector) []ConsumerRegistration {
	// Resolve shared dependencies from the DI container.
	db, err := do.Invoke[*bun.DB](injector)
	if err != nil {
		logger.Warnf("[queue] injector: failed to resolve *bun.DB: %v; notification log persistence disabled", err)
	}
	registry, err := do.Invoke[*notiftemplate.Registry](injector)
	if err != nil {
		logger.Warnf("[queue] injector: failed to resolve *notiftemplate.Registry: %v; template rendering disabled", err)
	}

	// Build renderer and log repo (nil-safe: if DB is unavailable, logs are skipped).
	var renderer *services.TemplateRenderer
	var logRepo *repositories.NotificationLogRepository
	if db != nil {
		overrideRepo := repositories.NewNotificationTemplateOverrideRepository(db)
		logRepo = repositories.NewNotificationLogRepository(db)
		if registry != nil {
			renderer = services.NewTemplateRenderer(registry, overrideRepo)
		}
	}

	// Resolve FCM-backed push channel (may be nil if FCM disabled).
	fcmClient, _ := do.Invoke[*fcm.Client](injector)
	pushChannel := notifChannels.NewPushChannel(fcmClient)

	// Resolve Redis client for idempotency guard (may be nil if Redis is unavailable).
	redisClient, _ := do.Invoke[*redis.Client](injector)

	// Build blast/user producers directly from the RabbitMQ connection.
	// This avoids importing the notification application package (which would create a cycle).
	conn, _ := do.Invoke[*rabbitmq.Connection](injector)
	blastProducer := buildExchangeProducer(conn, "notification.blast")
	userProducer := buildExchangeProducer(conn, "notification.user")

	// Shared channel set available to both blast and user consumers.
	chs := []notifChannels.NotificationChannel{
		notifChannels.NewEmailChannel(),
		pushChannel,
	}

	return []ConsumerRegistration{
		// Payment events consumer
		{
			Name:        "payment_handler",
			ConsumeFunc: orderConsumers.NewPaymentConsumer().Consume,
			Description: "Processes payment events (completed, failed, refunded)",
		},
		// Welcome notification consumer (legacy, kept for backward compatibility)
		{
			Name:        "welcome_notifier",
			ConsumeFunc: userConsumers.NewWelcomeNotificationConsumer().Consume,
			Description: "Sends welcome notifications to new users",
		},
		// Dispatcher: receives delayed messages from app.events and re-routes to blast/user exchanges.
		{
			Name: "notification_dispatcher",
			ConsumeFunc: notifConsumers.NewDispatcherConsumer(
				blastProducer,
				userProducer,
			).Consume,
			Description: "Routes delayed notification messages to the correct blast/user exchange",
		},
		// Blast: one publish → every user (fanout exchange)
		{
			Name:        "notification_blast",
			ConsumeFunc: notifConsumers.NewBlastConsumer(renderer, logRepo, chs...).Consume,
			Description: "Delivers broadcast notifications to all users via email and push",
		},
		// User-specific: one publish → one user (direct exchange, routing_key=user.<id>)
		{
			Name:        "notification_user",
			ConsumeFunc: notifConsumers.NewUserNotificationConsumer(renderer, logRepo, redisClient, chs...).Consume,
			Description: "Delivers targeted notifications to a single user via email and push",
		},
	}
}

// buildExchangeProducer creates a RabbitMQ producer bound to a specific exchange name.
// Returns nil when the connection is unavailable — callers handle nil gracefully.
func buildExchangeProducer(conn *rabbitmq.Connection, defaultExchangeName string) rabbitmq.MessageProducer {
	if conn == nil {
		return nil
	}

	exchangeName := viper.GetString("queue.rabbitmq.exchanges." + defaultExchangeName + ".name")
	if exchangeName == "" {
		exchangeName = defaultExchangeName
	}

	// Build a minimal config with just the publisher exchange set.
	cfg := rabbitmq.Config{
		Publisher: rabbitmq.PublisherConfig{
			ExchangeName: exchangeName,
		},
	}

	producer, err := rabbitmq.NewProducer(conn, cfg)
	if err != nil {
		logger.Warnf("[queue] buildExchangeProducer: failed to create producer for exchange %q: %v", defaultExchangeName, err)
		return nil
	}
	return producer
}
