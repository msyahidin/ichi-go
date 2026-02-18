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

	// userRoutingKeyPrefix is combined with userID for direct exchange routing.
	// The consumer queue must be bound with the same key pattern.
	userRoutingKeyPrefix = "user."
)

// NotificationService publishes notification events to RabbitMQ.
//
// It abstracts the two delivery modes behind clear method names so callers
// never need to know about exchanges, routing keys, or DeliveryMode values.
type NotificationService struct {
	producer             rabbitmq.MessageProducer
	blastExchangeName    string // fanout exchange — e.g. "notification.blast"
	userExchangeName     string // direct exchange — e.g. "notification.user"
}

func NewNotificationService(
	producer rabbitmq.MessageProducer,
	blastExchange string,
	userExchange string,
) *NotificationService {
	return &NotificationService{
		producer:          producer,
		blastExchangeName: blastExchange,
		userExchangeName:  userExchange,
	}
}

// Blast publishes a notification to ALL users via the fanout exchange.
// Every queue bound to the blast exchange receives a copy.
//
// The event's DeliveryMode is set automatically.
func (s *NotificationService) Blast(ctx context.Context, event dto.NotificationEvent) error {
	event.DeliveryMode = dto.DeliveryModeBlast

	return s.producer.Publish(ctx, blastRoutingKey, event, rabbitmq.PublishOptions{
		Headers: amqp.Table{
			"x-delivery-mode": string(dto.DeliveryModeBlast),
			"x-event-type":    event.EventType,
			"x-event-id":      event.EventID,
		},
	})
}

// SendToUser publishes a notification to a SINGLE user via the direct exchange.
// Routing key is constructed as "user.<userID>" — only the queue bound with
// that exact key receives the message.
//
// The event's DeliveryMode and UserID are set/validated automatically.
func (s *NotificationService) SendToUser(ctx context.Context, userID string, event dto.NotificationEvent) error {
	if userID == "" {
		return fmt.Errorf("notification: userID must not be empty for user-specific delivery")
	}

	event.DeliveryMode = dto.DeliveryModeUser
	event.UserID = userID

	routingKey := userRoutingKeyPrefix + userID

	return s.producer.Publish(ctx, routingKey, event, rabbitmq.PublishOptions{
		Headers: amqp.Table{
			"x-delivery-mode": string(dto.DeliveryModeUser),
			"x-event-type":    event.EventType,
			"x-event-id":      event.EventID,
			"x-user-id":       userID,
		},
	})
}
