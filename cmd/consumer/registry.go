package consumer

import (
	"ichi-go/internal/applications/order/handler"
	"ichi-go/internal/infra/messaging/rabbitmq/interfaces"
)

type Registration struct {
	Name    string
	Handler interfaces.MessageHandler
}

func RegisterConsumers() []Registration {
	return []Registration{
		//{"email_verifier", handler.NewEmailHandler().Handle},
		{"payment_handler", handler.NewPaymentHandler().Handle},
	}
}
