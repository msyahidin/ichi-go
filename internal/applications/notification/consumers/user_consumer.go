package consumers

import (
	"context"
	"encoding/json"
	"ichi-go/internal/applications/notification/channels"
	"ichi-go/internal/applications/notification/dto"
	"ichi-go/pkg/logger"
)

// UserNotificationConsumer processes notifications targeting a SINGLE user.
//
// Bound to: notification.user exchange (direct)
// Routing key: "user.<userID>"  — only this user's queue receives the message.
//
// Use for: OTPs, order updates, account alerts, password resets,
// personal recommendations — anything that must reach exactly one person.
//
// Flow:
//
//	NotificationService.SendToUser("usr_123", event)
//	  → publishes to notification.user (direct), routing_key="user.usr_123"
//	  → notification.user.usr_123.queue  → UserNotificationConsumer
//
// Note: For high-volume systems, consider pre-creating queues per user on
// first login and binding them here. For lower volume, a single shared queue
// with routing_key="#" and per-message userID filtering also works.
type UserNotificationConsumer struct {
	channels []channels.NotificationChannel
}

func NewUserNotificationConsumer(chs ...channels.NotificationChannel) *UserNotificationConsumer {
	return &UserNotificationConsumer{channels: chs}
}

// Consume is the ConsumeFunc registered in registry.go.
func (c *UserNotificationConsumer) Consume(ctx context.Context, body []byte) error {
	var event dto.NotificationEvent
	if err := json.Unmarshal(body, &event); err != nil {
		logger.Errorf("[user-notif] invalid JSON, discarding: %v", err)
		return nil
	}

	if event.DeliveryMode != dto.DeliveryModeUser {
		logger.Warnf("[user-notif] unexpected delivery_mode=%s event_id=%s, discarding",
			event.DeliveryMode, event.EventID)
		return nil
	}

	if event.UserID == "" {
		logger.Errorf("[user-notif] missing user_id event_id=%s, discarding", event.EventID)
		return nil
	}

	logger.Infof("[user-notif] dispatching event_id=%s event_type=%s user_id=%s channels=%v",
		event.EventID, event.EventType, event.UserID, event.Channels)

	return dispatch(ctx, event, c.channels)
}
