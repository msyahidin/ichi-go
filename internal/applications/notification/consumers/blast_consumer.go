package consumers

import (
	"context"
	"encoding/json"
	"ichi-go/internal/applications/notification/channels"
	"ichi-go/internal/applications/notification/dto"
	"ichi-go/pkg/logger"
)

// BlastConsumer processes broadcast notifications sent to ALL users.
//
// Bound to: notification.blast exchange (fanout)
// Routing key: ignored by fanout — every bound queue receives the message.
//
// Use for: system announcements, maintenance windows, feature releases,
// promotional campaigns targeting every user.
//
// Flow:
//
//	NotificationService.Blast(event)
//	  → publishes to notification.blast (fanout)
//	  → notification.blast.email.queue  → BlastConsumer (email channel only)
//	  → notification.blast.push.queue   → BlastConsumer (push channel only)
//	  → notification.blast.sms.queue    → BlastConsumer (sms channel only)
//
// Each queue can have independent worker counts and retry budgets.
type BlastConsumer struct {
	channels []channels.NotificationChannel
}

func NewBlastConsumer(chs ...channels.NotificationChannel) *BlastConsumer {
	return &BlastConsumer{channels: chs}
}

// Consume is the ConsumeFunc registered in registry.go.
func (c *BlastConsumer) Consume(ctx context.Context, body []byte) error {
	var event dto.NotificationEvent
	if err := json.Unmarshal(body, &event); err != nil {
		// Bad JSON is a permanent failure — ack and discard, never retry.
		logger.Errorf("[blast] invalid JSON, discarding: %v", err)
		return nil
	}

	if event.DeliveryMode != dto.DeliveryModeBlast {
		logger.Warnf("[blast] unexpected delivery_mode=%s event_id=%s, discarding",
			event.DeliveryMode, event.EventID)
		return nil
	}

	logger.Infof("[blast] dispatching event_id=%s event_type=%s channels=%v",
		event.EventID, event.EventType, event.Channels)

	return dispatch(ctx, event, c.channels)
}
