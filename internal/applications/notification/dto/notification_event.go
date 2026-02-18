package dto

// DeliveryMode identifies which exchange strategy to use when publishing.
type DeliveryMode string

const (
	// DeliveryModeBlast sends to ALL users via fanout exchange.
	// Use for system announcements, maintenance notices, feature releases.
	DeliveryModeBlast DeliveryMode = "blast"

	// DeliveryModeUser sends to a SINGLE user via direct exchange.
	// Use for personal alerts, OTPs, order updates, account events.
	DeliveryModeUser DeliveryMode = "user"
)

// Channel identifies a notification delivery channel.
type Channel string

const (
	ChannelEmail   Channel = "email"
	ChannelPush    Channel = "push"
	ChannelSMS     Channel = "sms"
	ChannelInApp   Channel = "in_app"
	ChannelWebhook Channel = "webhook"
)

// NotificationEvent is the unified envelope for all notification messages.
//
// Blast example:
//
//	NotificationEvent{
//	    EventID:      uuid.New().String(),
//	    EventType:    "system.maintenance",
//	    DeliveryMode: DeliveryModeBlast,
//	    Channels:     []Channel{ChannelEmail, ChannelPush},
//	    Data:         map[string]any{"window": "2026-02-17 02:00–04:00 UTC"},
//	}
//
// User-specific example:
//
//	NotificationEvent{
//	    EventID:      uuid.New().String(),
//	    EventType:    "order.shipped",
//	    DeliveryMode: DeliveryModeUser,
//	    UserID:       "usr_123",
//	    Channels:     []Channel{ChannelEmail, ChannelPush},
//	    Data:         map[string]any{"order_id": "ord_456", "tracking": "JNE-789"},
//	}
type NotificationEvent struct {
	// EventID is a unique identifier for idempotent processing.
	// Use UUID. Consumers should skip duplicate EventIDs.
	EventID string `json:"event_id"`

	// EventType is a dot-separated string describing what happened.
	// Used by channel consumers to select the correct template.
	// Examples: "user.welcome", "order.shipped", "system.maintenance"
	EventType string `json:"event_type"`

	// DeliveryMode controls which exchange routes this message.
	DeliveryMode DeliveryMode `json:"delivery_mode"`

	// Channels lists which channels should deliver this notification.
	// Channel consumers check this list and skip if not included.
	Channels []Channel `json:"channels"`

	// UserID is the target user. Required when DeliveryMode is DeliveryModeUser.
	// Empty for blast notifications.
	UserID string `json:"user_id,omitempty"`

	// Locale is the BCP-47 language tag for template selection.
	// Defaults to "en" if empty.
	Locale string `json:"locale,omitempty"`

	// Data holds template variables for rendering the notification body.
	// Keys depend on the EventType and channel template.
	Data map[string]any `json:"data,omitempty"`

	// Meta holds operational metadata: tenant_id, trace_id, source service.
	// Not shown to end users.
	Meta map[string]string `json:"meta,omitempty"`
}

// HasChannel reports whether the event targets a specific channel.
func (e *NotificationEvent) HasChannel(ch Channel) bool {
	for _, c := range e.Channels {
		if c == ch {
			return true
		}
	}
	return false
}
