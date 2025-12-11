package messaging

import (
	"ichi-go/internal/applications/order/handler"
	"ichi-go/internal/infra/messaging/rabbitmq"
)

type Registration struct {
	Name    string
	Handler rabbitmq.MessageHandler
}

func GetRegisteredConsumers() []Registration {
	return []Registration{
		//{"email_verifier", handler.NewEmailHandler().Handle},
		{"payment_handler", handler.NewPaymentHandler().Handle},
	}
}
