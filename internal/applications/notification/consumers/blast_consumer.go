package consumers

import (
	"context"
	"encoding/json"
	"strconv"

	"ichi-go/internal/applications/notification/channels"
	"ichi-go/internal/applications/notification/dto"
	"ichi-go/internal/applications/notification/repositories"
	"ichi-go/internal/applications/notification/services"
	"ichi-go/pkg/logger"
)

// BlastConsumer processes broadcast notifications sent to ALL users.
//
// Bound to: notification.blast exchange (fanout)
// Routing key: ignored by fanout — every bound queue receives the message.
//
// Use for: system announcements, maintenance windows, feature releases,
// promotional campaigns targeting every user.
type BlastConsumer struct {
	channels []channels.NotificationChannel
	renderer *services.TemplateRenderer
	logRepo  *repositories.NotificationLogRepository
}

func NewBlastConsumer(
	renderer *services.TemplateRenderer,
	logRepo *repositories.NotificationLogRepository,
	chs ...channels.NotificationChannel,
) *BlastConsumer {
	return &BlastConsumer{
		channels: chs,
		renderer: renderer,
		logRepo:  logRepo,
	}
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

	// Extract campaign_id from meta for log correlation.
	campaignID := extractCampaignID(event.Meta)

	return dispatch(ctx, event, c.channels, c.renderer, c.logRepo, campaignID)
}

// extractCampaignID reads the campaign_id from the event's meta map.
// Returns 0 if not present (logs will still be written, just without campaign linkage).
func extractCampaignID(meta map[string]string) int64 {
	if meta == nil {
		return 0
	}
	if idStr, ok := meta["campaign_id"]; ok {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		return id
	}
	return 0
}
