package channels

import (
	"context"
	"ichi-go/internal/applications/notification/dto"
)

// NotificationChannel is implemented by each delivery channel (email, push, SMS, etc).
//
// Adding a new channel:
//  1. Create a new file in this package, e.g. slack_channel.go
//  2. Implement this interface
//  3. Register in the consumer's channel list via DI
//
// No existing consumer code changes are needed.
type NotificationChannel interface {
	// Name returns the channel identifier, matching dto.Channel constants.
	Name() dto.Channel

	// Send delivers the notification via this channel.
	//
	// Return error only for transient failures (network, rate limit).
	// Return nil for permanent failures (invalid address, user opted out) —
	// these should be logged but not retried.
	Send(ctx context.Context, event dto.NotificationEvent) error
}
