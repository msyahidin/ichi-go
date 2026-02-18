package queue

import (
	orderConsumers "ichi-go/internal/applications/order/consumers"
	"ichi-go/internal/applications/notification/channels"
	notifConsumers "ichi-go/internal/applications/notification/consumers"
	userConsumers "ichi-go/internal/applications/user/consumers"
	"ichi-go/internal/infra/queue/rabbitmq"
)

// ConsumerRegistration links consumer name to processing function.
type ConsumerRegistration struct {
	Name        string               // Must match config.yaml
	ConsumeFunc rabbitmq.ConsumeFunc // Processing function
	Description string               // What this consumer does
}

// GetRegisteredConsumers returns all queue consumers.
//
// To add new consumer:
// 1. Create in internal/applications/{domain}/consumers/
// 2. Implement Consume(ctx, body) error
// 3. Add registration here
// 4. Add config in config.yaml
// 5. Test
func GetRegisteredConsumers() []ConsumerRegistration {
	// Shared channel instances — add new channels (SMS, Slack, webhook) here.
	// Both blast and user consumers receive the same set so all channels
	// are available for both delivery modes.
	notifChannels := []channels.NotificationChannel{
		channels.NewEmailChannel(),
		channels.NewPushChannel(),
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
		// Blast: one publish → every user (fanout exchange)
		{
			Name:        "notification_blast",
			ConsumeFunc: notifConsumers.NewBlastConsumer(notifChannels...).Consume,
			Description: "Delivers broadcast notifications to all users via email and push",
		},
		// User-specific: one publish → one user (direct exchange, routing_key=user.<id>)
		{
			Name:        "notification_user",
			ConsumeFunc: notifConsumers.NewUserNotificationConsumer(notifChannels...).Consume,
			Description: "Delivers targeted notifications to a single user via email and push",
		},
	}
}
