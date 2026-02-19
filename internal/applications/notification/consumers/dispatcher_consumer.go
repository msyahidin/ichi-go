package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"ichi-go/internal/applications/notification/dto"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/logger"
)

const (
	// blastRoutingKey is the routing key for the fanout exchange.
	// Fanout ignores routing keys, but we set one for observability.
	blastRoutingKey = "notification.blast"

	// userRoutingKeyPrefix is prepended to userID for direct exchange routing.
	userRoutingKeyPrefix = "user."
)

// DispatcherConsumer is bound to the app.events (x-delayed-message) exchange
// with routing key "notification.dispatch".
//
// Its sole responsibility is to re-route delayed notification messages to the
// correct exchange (blast fanout OR user direct) with zero delay.
//
// Why a dispatcher instead of publishing directly to blast/user exchanges?
// The blast and user exchanges are NOT x-delayed-message type. All delay logic
// lives in one place (the app.events exchange), keeping blast/user exchanges clean.
//
// Flow:
//
//	CampaignService → app.events (x-delay=N) → [delay expires] → DispatcherConsumer
//	  → blast: blastProducer.Publish("notification.blast", event)
//	  → user:  userProducer.Publish("user.<userID>", event)
type DispatcherConsumer struct {
	blastProducer rabbitmq.MessageProducer // bound to notification.blast (fanout) exchange
	userProducer  rabbitmq.MessageProducer // bound to notification.user (direct) exchange
}

func NewDispatcherConsumer(
	blastProducer rabbitmq.MessageProducer,
	userProducer rabbitmq.MessageProducer,
) *DispatcherConsumer {
	return &DispatcherConsumer{
		blastProducer: blastProducer,
		userProducer:  userProducer,
	}
}

// Consume processes a delayed notification message and re-routes it.
func (c *DispatcherConsumer) Consume(ctx context.Context, body []byte) error {
	var event dto.NotificationEvent
	if err := json.Unmarshal(body, &event); err != nil {
		// Bad JSON is a permanent failure — ack and discard, never retry.
		logger.Errorf("[dispatcher] invalid JSON, discarding: %v", err)
		return nil
	}

	logger.Infof("[dispatcher] routing event_id=%s event_type=%s delivery_mode=%s",
		event.EventID, event.EventType, event.DeliveryMode)

	opts := rabbitmq.PublishOptions{
		Headers: amqp.Table{
			"x-event-type":    event.EventType,
			"x-event-id":      event.EventID,
			"x-delivery-mode": string(event.DeliveryMode),
		},
		// No delay on re-publish — deliver immediately.
	}

	switch event.DeliveryMode {
	case dto.DeliveryModeBlast:
		if err := c.blastProducer.Publish(ctx, blastRoutingKey, event, opts); err != nil {
			logger.Errorf("[dispatcher] blast re-publish failed event_id=%s: %v", event.EventID, err)
			return err // transient — requeue for retry
		}
		logger.Debugf("[dispatcher] blast routed event_id=%s", event.EventID)
		return nil

	case dto.DeliveryModeUser:
		if event.UserID == "" {
			logger.Errorf("[dispatcher] user delivery_mode but empty user_id event_id=%s, discarding", event.EventID)
			return nil // permanent failure — discard
		}
		routingKey := userRoutingKeyPrefix + event.UserID
		if err := c.userProducer.Publish(ctx, routingKey, event, opts); err != nil {
			logger.Errorf("[dispatcher] user re-publish failed event_id=%s user_id=%s: %v",
				event.EventID, event.UserID, err)
			return err // transient — requeue for retry
		}
		logger.Debugf("[dispatcher] user routed event_id=%s user_id=%s", event.EventID, event.UserID)
		return nil

	default:
		logger.Errorf("[dispatcher] unknown delivery_mode=%q event_id=%s, discarding",
			event.DeliveryMode, event.EventID)
		return fmt.Errorf("unknown delivery_mode: %s", event.DeliveryMode) // won't retry if consumer is configured correctly
	}
}
