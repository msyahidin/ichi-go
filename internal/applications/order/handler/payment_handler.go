package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"ichi-go/pkg/logger"
)

type PaymentEvent struct {
	EventType string  `json:"event_type"`
	UserID    int64   `json:"user_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

type PaymentHandler struct {
	// Add dependencies
}

func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{}
}

func (h *PaymentHandler) Handle(ctx context.Context, body []byte) error {
	var event PaymentEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	logger.Debugf("💳 Processing payment event: %s for user %d (amount: %.2f)",
		event.EventType, event.UserID, event.Amount)

	switch event.EventType {
	case "payment.completed":
		logger.Debugf("✅ Payment completed: $%.2f", event.Amount)

	case "payment.failed":
		logger.Debugf("❌ Payment failed for user %d", event.UserID)

	default:
		if len(event.EventType) > 13 && event.EventType[:13] == "subscription." {
			logger.Debugf("🔄 Subscription event: %s", event.EventType)
		}
	}

	return nil
}
