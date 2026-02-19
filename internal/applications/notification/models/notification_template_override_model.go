package models

import (
	"ichi-go/pkg/db/model"
)

// NotificationTemplateOverride stores optional runtime copy overrides for
// notification templates. Rows here take precedence over Go-defined defaults.
//
// The Go TemplateRegistry (pkg/notification/template) is the source of truth.
// Missing rows are not errors — the system falls back to Go defaults.
type NotificationTemplateOverride struct {
	model.CoreModel `bun:"table:notification_template_overrides,alias:nto"`

	// EventSlug must match a slug registered in the Go TemplateRegistry.
	// There is no FK to a DB events table — the registry is in code.
	EventSlug string `bun:"event_slug,notnull" json:"event_slug"`

	// Channel identifies the delivery channel.
	// Values: email | push | sms | in_app | webhook
	Channel string `bun:"channel,notnull" json:"channel"`

	// Locale is the BCP-47 language tag. Default: "en"
	Locale string `bun:"locale,notnull,default:'en'" json:"locale"`

	// TitleTemplate is an optional Go text/template string for the notification title.
	// When non-empty, overrides the Go template's DefaultContent().Title.
	TitleTemplate string `bun:"title_template" json:"title_template,omitempty"`

	// BodyTemplate is an optional Go text/template string for the notification body.
	// When non-empty, overrides the Go template's DefaultContent().Body.
	BodyTemplate string `bun:"body_template" json:"body_template,omitempty"`

	// IsActive controls whether this override is used. Inactive rows are ignored.
	IsActive bool `bun:"is_active,notnull,default:true" json:"is_active"`
}
