package consumers

import (
	"context"
	"encoding/json"
	"ichi-go/internal/applications/user/dto"
	"ichi-go/pkg/logger"
)

// WelcomeNotificationConsumer processes welcome notification messages
type WelcomeNotificationConsumer struct {
	// Add dependencies for email service, notification service, etc.
	// emailService email.Service
	// notificationService notification.Service
}

func NewWelcomeNotificationConsumer() *WelcomeNotificationConsumer {
	return &WelcomeNotificationConsumer{}
}

// Consume processes welcome notification message
// Return error for transient failures (will retry)
// Return nil for permanent failures (will skip)
func (c *WelcomeNotificationConsumer) Consume(ctx context.Context, body []byte) error {
	var message dto.WelcomeNotificationMessage

	// Parse message
	if err := json.Unmarshal(body, &message); err != nil {
		logger.Errorf("Invalid JSON - skipping: %v", err)
		return nil // Don't retry bad JSON
	}

	logger.Infof("📧 Processing welcome notification for user %s", message.UserID)

	// TODO: Send actual email
	if err := c.sendWelcomeEmail(ctx, message); err != nil {
		// Check if transient error (network, service down)
		if isTransientError(err) {
			logger.Errorf("Transient error - will retry: %v", err)
			return err // Retry
		}

		// Permanent error (invalid email, user deleted)
		logger.Errorf("Permanent error - skipping: %v", err)
		return nil // Don't retry
	}

	// TODO: Send push notification
	if err := c.sendPushNotification(ctx, message); err != nil {
		logger.Warnf("Push notification failed (non-critical): %v", err)
		// Don't fail the whole operation
	}

	logger.Infof("✅ Welcome notification sent to user %s", message.UserID)
	return nil
}

func (c *WelcomeNotificationConsumer) sendWelcomeEmail(ctx context.Context, msg dto.WelcomeNotificationMessage) error {
	// TODO: Implement email sending
	// emailService.Send(ctx, msg.Email, "Welcome!", msg.Text)

	logger.Debugf("📧 Would send email to %s: %s", msg.Email, msg.Text)
	return nil
}

func (c *WelcomeNotificationConsumer) sendPushNotification(ctx context.Context, msg dto.WelcomeNotificationMessage) error {
	// TODO: Implement push notification
	logger.Debugf("🔔 Would send push to user %s", msg.UserID)
	return nil
}

func isTransientError(err error) bool {
	// Check for network errors, timeouts, service unavailable, etc.
	// Example: return errors.Is(err, context.DeadlineExceeded)
	return false
}
