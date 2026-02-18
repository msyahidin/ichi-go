package consumers

import (
	"context"
	"encoding/json"

	"ichi-go/internal/applications/notification/channels"
	"ichi-go/internal/applications/notification/dto"
	"ichi-go/internal/applications/notification/repositories"
	"ichi-go/internal/applications/notification/services"
	"ichi-go/pkg/logger"
)

// UserNotificationConsumer processes notifications targeting a SINGLE user.
//
// Bound to: notification.user exchange (topic)
// Routing key pattern: "user.*"  — matches per-user keys like "user.42".
//
// Use for: OTPs, order updates, account alerts, password resets,
// personal recommendations — anything that must reach exactly one person.
type UserNotificationConsumer struct {
	channels []channels.NotificationChannel
	renderer *services.TemplateRenderer
	logRepo  *repositories.NotificationLogRepository
}

func NewUserNotificationConsumer(
	renderer *services.TemplateRenderer,
	logRepo *repositories.NotificationLogRepository,
	chs ...channels.NotificationChannel,
) *UserNotificationConsumer {
	return &UserNotificationConsumer{
		channels: chs,
		renderer: renderer,
		logRepo:  logRepo,
	}
}

// maskUserID returns a redacted form of the user ID suitable for logging.
// Shows only the last 3 characters prefixed with "***" so logs remain
// debuggable without exposing the full identifier.
// Example: "1234567" → "***567", "42" → "***42"
func maskUserID(uid string) string {
	const keepLast = 3
	if len(uid) <= keepLast {
		return "***" + uid
	}
	return "***" + uid[len(uid)-keepLast:]
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

	maskedUID := maskUserID(event.UserID)
	logger.Infof("[user-notif] dispatching event_id=%s event_type=%s user_id=%s channels=%v",
		event.EventID, event.EventType, maskedUID, event.Channels)

	campaignID := extractCampaignID(event.Meta)

	return dispatch(ctx, event, c.channels, c.renderer, c.logRepo, campaignID)
}
