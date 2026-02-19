package builtin

import (
	notiftemplate "ichi-go/pkg/notification/template"
)

// OrderShippedTemplate is the built-in template for the "order.shipped" event.
// Sent when an order has been dispatched from the warehouse.
//
// Required data variables:
//   - order_id   string  — order reference number
//   - name       string  — customer display name
//   - eta        string  — estimated arrival date/time string
//   - tracking_url string — carrier tracking link
type OrderShippedTemplate struct{}

func (t OrderShippedTemplate) Slug() string { return "order.shipped" }

func (t OrderShippedTemplate) SupportedChannels() []string {
	return []string{"email", "push", "sms"}
}

func (t OrderShippedTemplate) DefaultContent(channel, locale string) notiftemplate.ChannelContent {
	switch locale {
	case "id": // Bahasa Indonesia
		return t.indonesian(channel)
	default: // "en" and all other locales fall back to English
		return t.english(channel)
	}
}

func (t OrderShippedTemplate) english(channel string) notiftemplate.ChannelContent {
	switch channel {
	case "push":
		return notiftemplate.ChannelContent{
			Title: "Your order #{{.order_id}} is on its way!",
			Body:  "Estimated arrival: {{.eta}}",
		}
	case "sms":
		return notiftemplate.ChannelContent{
			Title: "",
			Body:  "Hi {{.name}}, your order #{{.order_id}} has shipped! Track it: {{.tracking_url}}",
		}
	default: // email
		return notiftemplate.ChannelContent{
			Title: "Order Shipped — #{{.order_id}}",
			Body:  "Hi {{.name}}, your order has shipped. Track it here: {{.tracking_url}}. Estimated arrival: {{.eta}}.",
		}
	}
}

func (t OrderShippedTemplate) indonesian(channel string) notiftemplate.ChannelContent {
	switch channel {
	case "push":
		return notiftemplate.ChannelContent{
			Title: "Pesanan #{{.order_id}} sudah dikirim!",
			Body:  "Estimasi tiba: {{.eta}}",
		}
	case "sms":
		return notiftemplate.ChannelContent{
			Title: "",
			Body:  "Hai {{.name}}, pesanan #{{.order_id}} sudah dikirim! Lacak: {{.tracking_url}}",
		}
	default: // email
		return notiftemplate.ChannelContent{
			Title: "Pesanan Dikirim — #{{.order_id}}",
			Body:  "Hai {{.name}}, pesanan Anda sudah dikirim. Lacak di sini: {{.tracking_url}}. Estimasi tiba: {{.eta}}.",
		}
	}
}

func init() {
	notiftemplate.GlobalRegistry.Register(OrderShippedTemplate{})
}
