package template

// EventTemplate is the Go-code contract every notification event must implement.
//
// Adding a new notification event:
//  1. Create a new file in pkg/notification/template/builtin/, e.g. order_shipped.go
//  2. Implement this interface
//  3. Register via GlobalRegistry.Register() in an init() function
//
// The template defines:
//   - Which channels this event supports
//   - Default title + body copy per channel and locale
//
// The DB table notification_template_overrides can override the copy at runtime
// without a redeploy. If no DB override exists, DefaultContent() is used.
type EventTemplate interface {
	// Slug returns the unique event identifier used in API calls and DB lookups.
	// Use dot-separated notation: "order.shipped", "user.welcome", "system.maintenance"
	Slug() string

	// SupportedChannels returns the list of channel identifiers this event can be
	// delivered through. API validation rejects channels not in this list.
	// Values should match dto.Channel constants: "email", "push", "sms", "in_app", "webhook"
	SupportedChannels() []string

	// DefaultContent returns the fallback title and body template strings for the
	// given channel and locale. Called when no DB override row exists.
	//
	// Templates use Go text/template syntax: "Hello {{.name}}, your order {{.order_id}} shipped!"
	// Use option("missingkey=zero") — missing keys render as empty string, not panic.
	//
	// channel: matches dto.Channel constants
	// locale:  BCP-47 tag, e.g. "en", "id"
	DefaultContent(channel, locale string) ChannelContent
}

// ChannelContent holds the rendered title and body template strings for a
// specific channel and locale combination.
type ChannelContent struct {
	// Title is the notification subject/title Go text/template string.
	// For email: the email subject line.
	// For push: the push notification title.
	// For SMS: not used (body only).
	Title string

	// Body is the notification body Go text/template string.
	// For email: plain text body (HTML templates belong in a separate email template file).
	// For push: the notification body text.
	// For SMS: the full message text.
	Body string
}
