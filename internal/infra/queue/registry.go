package queue

import (
	"ichi-go/internal/applications/order/consumers"
	"ichi-go/internal/infra/queue/rabbitmq"
)

// ConsumerRegistration links consumer name to processing function.
type ConsumerRegistration struct {
	Name        string               // Must match config.yaml
	ConsumeFunc rabbitmq.ConsumeFunc // Processing function
	Description string               // What this consumer does
}

// GetRegisteredConsumers returns all queue consumers.
//
// To add new consumer:
// 1. Create in internal/applications/{domain}/consumers/
// 2. Implement Consume(ctx, body) error
// 3. Add registration here
// 4. Add config in config.yaml
// 5. Test
func GetRegisteredConsumers() []ConsumerRegistration {
	return []ConsumerRegistration{
		{
			Name:        "payment_handler",
			ConsumeFunc: consumers.NewPaymentConsumer().Consume,
			Description: "Processes payment events (completed, failed, refunded)",
		},
	}
}
