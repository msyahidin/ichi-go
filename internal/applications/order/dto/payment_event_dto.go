package dto

// PaymentCompletedEvent represents a successful payment
type PaymentCompletedEvent struct {
	EventType     string  `json:"event_type"` // "payment.completed"
	OrderID       string  `json:"order_id"`
	UserID        int64   `json:"user_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
	TransactionID string  `json:"transaction_id"`
}

// PaymentFailedEvent represents a failed payment
type PaymentFailedEvent struct {
	EventType string  `json:"event_type"` // "payment.failed"
	OrderID   string  `json:"order_id"`
	UserID    int64   `json:"user_id"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
	ErrorCode string  `json:"error_code"`
}

// PaymentRefundedEvent represents a refund
type PaymentRefundedEvent struct {
	EventType string  `json:"event_type"` // "payment.refunded"
	OrderID   string  `json:"order_id"`
	UserID    int64   `json:"user_id"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
	RefundID  string  `json:"refund_id"`
}
