package channels

import (
	"context"
	"ichi-go/internal/applications/notification/dto"
	"ichi-go/pkg/logger"
)

// PushChannel delivers notifications via mobile/web push (FCM, APNs, etc).
type PushChannel struct {
	// pushClient push.Client  — inject your real client here
}

func NewPushChannel() *PushChannel {
	return &PushChannel{}
}

func (c *PushChannel) Name() dto.Channel {
	return dto.ChannelPush
}

func (c *PushChannel) Send(ctx context.Context, event dto.NotificationEvent) error {
	// TODO: look up device tokens for event.UserID
	// TODO: resolve push payload by event.EventType
	// TODO: call c.pushClient.Send(ctx, tokens, payload)
	logger.Infof("[push] sending event_type=%s user_id=%s", event.EventType, event.UserID)
	return nil
}
