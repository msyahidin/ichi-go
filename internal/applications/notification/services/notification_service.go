package services

import (
	"context"
	"fmt"
	"ichi-go/internal/applications/notification/dto"
	"ichi-go/internal/infra/queue/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// blastRoutingKey is published to the fanout exchange.
	// Fanout ignores routing keys, but we set one for observability/tracing.
	blastRoutingKey = "notification.blast"

	// userRoutingKeyPrefix is combined with userID for topic exchange routing.
	// The consumer queue must be bound with the pattern "user.#".
	userRoutingKeyPrefix = "user."
)

// NotificationService publishes notification events to RabbitMQ.
//
// It abstracts the two delivery modes behind clear method names so callers
// never need to know about exchanges, routing keys, or DeliveryMode values.
//
// blastProducer must be bound to the fanout blast exchange.
// userProducer must be bound to the topic user exchange.
type NotificationService struct {
	blastProducer rabbitmq.MessageProducer // fanout exchange — notification.blast
	userProducer  rabbitmq.MessageProducer // topic exchange  — notification.user
}

func NewNotificationService(
	blastProducer rabbitmq.MessageProducer,
	userProducer rabbitmq.MessageProducer,
) *NotificationService {
	return &NotificationService{
		blastProducer: blastProducer,
		userProducer:  userProducer,
	}
}

// Blast publishes a notification to ALL users via the fanout exchange.
// Every queue bound to the blast exchange receives a copy.
//
// The event's DeliveryMode is set automatically.
// EventID must be non-empty to ensure idempotency headers are useful.
func (s *NotificationService) Blast(ctx context.Context, event dto.NotificationEvent) error {
	if s.blastProducer == nil {
		return fmt.Errorf("notification: blast producer unavailable")
	}
	if event.EventID == "" {
		return fmt.Errorf("notification: EventID must not be empty (required for idempotency)")
	}

	event.DeliveryMode = dto.DeliveryModeBlast

	return s.blastProducer.Publish(ctx, blastRoutingKey, event, rabbitmq.PublishOptions{
		Headers: amqp.Table{
			"x-delivery-mode": string(dto.DeliveryModeBlast),
			"x-event-type":    event.EventType,
			"x-event-id":      event.EventID,
		},
	})
}

// SendToUser publishes a notification to a SINGLE user via the topic exchange.
// Routing key is constructed as "user.<userID>" — only the queue bound with
// the pattern "user.#" on the topic exchange receives the message.
//
// The event's DeliveryMode and UserID are set/validated automatically.
// EventID must be non-empty to ensure idempotency headers are useful.
func (s *NotificationService) SendToUser(ctx context.Context, userID string, event dto.NotificationEvent) error {
	if s.userProducer == nil {
		return fmt.Errorf("notification: user producer unavailable")
	}
	if userID == "" {
		return fmt.Errorf("notification: userID must not be empty for user-specific delivery")
	}
	if event.EventID == "" {
		return fmt.Errorf("notification: EventID must not be empty (required for idempotency)")
	}

	event.DeliveryMode = dto.DeliveryModeUser
	event.UserID = userID

	routingKey := userRoutingKeyPrefix + userID

	return s.userProducer.Publish(ctx, routingKey, event, rabbitmq.PublishOptions{
		Headers: amqp.Table{
			"x-delivery-mode": string(dto.DeliveryModeUser),
			"x-event-type":    event.EventType,
			"x-event-id":      event.EventID,
			"x-user-id":       userID,
		},
	})
}
