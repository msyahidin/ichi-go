package consumers

import (
	"context"
	"encoding/json"
	"ichi-go/pkg/logger"
)

// PaymentEvent represents payment message.
type PaymentEvent struct {
	EventType string  `json:"event_type"`
	UserID    int64   `json:"user_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

// PaymentConsumer processes payment events.
//
// IMPORTANT: Queue consumer, NOT HTTP handler!
//
// Handles:
// - payment.completed
// - payment.failed
// - payment.refunded
type PaymentConsumer struct {
	// TODO: Add dependencies
}

// NewPaymentConsumer creates consumer.
func NewPaymentConsumer() *PaymentConsumer {
	return &PaymentConsumer{}
}

// Consume processes payment message.
//
// Error Handling:
// - Return error for transient failures (retry)
// - Return nil for permanent failures (skip)
func (c *PaymentConsumer) Consume(ctx context.Context, body []byte) error {
	var event PaymentEvent
	if err := json.Unmarshal(body, &event); err != nil {
		logger.Errorf("Invalid message: %v", err)
		return nil // Don't retry bad JSON
	}

	logger.Infof("💳 Processing: type=%s, user=%d, amount=%.2f",
		event.EventType, event.UserID, event.Amount)

	switch event.EventType {
	case "payment.completed":
		return c.handleCompleted(ctx, event)
	case "payment.failed":
		return c.handleFailed(ctx, event)
	case "payment.refunded":
		return c.handleRefunded(ctx, event)
	default:
		logger.Warnf("Unknown event: %s", event.EventType)
		return nil
	}
}

// handleCompleted processes successful payments.
func (c *PaymentConsumer) handleCompleted(ctx context.Context, event PaymentEvent) error {
	logger.Debugf("✅ Completed: $%.2f for user %d", event.Amount, event.UserID)

	// TODO: Implement
	// - Update order status
	// - Send confirmation
	// - Trigger fulfillment

	logger.Infof("✅ Payment completed: user %d", event.UserID)
	return nil
}

// handleFailed processes failed payments.
func (c *PaymentConsumer) handleFailed(ctx context.Context, event PaymentEvent) error {
	logger.Debugf("❌ Failed: user %d", event.UserID)

	// TODO: Implement
	// - Update order status
	// - Send notification
	// - Log analytics

	logger.Warnf("❌ Payment failed: user %d", event.UserID)
	return nil
}

// handleRefunded processes refunds.
func (c *PaymentConsumer) handleRefunded(ctx context.Context, event PaymentEvent) error {
	logger.Debugf("🔄 Refunded: $%.2f for user %d", event.Amount, event.UserID)

	// TODO: Implement
	// - Update order status
	// - Return inventory
	// - Send confirmation

	logger.Infof("🔄 Refunded: user %d", event.UserID)
	return nil
}
