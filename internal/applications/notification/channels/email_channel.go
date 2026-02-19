package channels

import (
	"context"
	"ichi-go/internal/applications/notification/dto"
	"ichi-go/pkg/logger"
)

// EmailChannel delivers notifications via email.
// Replace the Send body with your real email client (SES, SendGrid, SMTP, etc).
type EmailChannel struct {
	// emailClient email.Client  — inject your real client here
}

func NewEmailChannel() *EmailChannel {
	return &EmailChannel{}
}

func (c *EmailChannel) Name() dto.Channel {
	return dto.ChannelEmail
}

func (c *EmailChannel) Send(ctx context.Context, event dto.NotificationEvent) error {
	// TODO: look up user email from event.UserID or event.Data["email"]
	// TODO: resolve template by event.EventType + event.Locale
	// TODO: call c.emailClient.Send(ctx, to, subject, body)
	logger.Infof("[email] sending event_type=%s user_id=%s", event.EventType, event.UserID)
	return nil
}
