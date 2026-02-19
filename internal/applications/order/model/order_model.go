package model

import (
	"context"
	"fmt"
	"github.com/uptrace/bun"
	"ichi-go/pkg/db/model"
	"time"
)

// Order represents a customer order in the system.
// This model supports the full order lifecycle: creation, payment, fulfillment, and refunds.
type Order struct {
	model.CoreModel `bun:"table:orders,alias:o"`

	// Order Identification
	OrderNumber string `bun:"order_number,notnull,unique" json:"order_number"` // Human-readable order number (e.g., "ORD-20250101-0001")

	// Customer Information
	UserID   int64  `bun:"user_id,notnull" json:"user_id"`
	UserName string `bun:"user_name,notnull" json:"user_name"`
	Email    string `bun:"email,notnull" json:"email"`
	Phone    string `bun:"phone" json:"phone"`

	// Order Status
	// Possible values: "pending", "payment_pending", "paid", "payment_failed",
	// "processing", "shipped", "delivered", "cancelled", "refunded"
	Status string `bun:"status,notnull,default:'pending'" json:"status"`

	// Financial Information
	SubtotalAmount   float64 `bun:"subtotal_amount,type:decimal(15,2),notnull,default:0" json:"subtotal_amount"`     // Sum of all items
	TaxAmount        float64 `bun:"tax_amount,type:decimal(15,2),notnull,default:0" json:"tax_amount"`               // Tax
	ShippingAmount   float64 `bun:"shipping_amount,type:decimal(15,2),notnull,default:0" json:"shipping_amount"`     // Shipping cost
	DiscountAmount   float64 `bun:"discount_amount,type:decimal(15,2),notnull,default:0" json:"discount_amount"`     // Applied discounts
	TotalAmount      float64 `bun:"total_amount,type:decimal(15,2),notnull,default:0" json:"total_amount"`           // Final amount to pay
	Currency         string  `bun:"currency,notnull,default:'IDR'" json:"currency"`                                  // Currency code
	RefundedAmount   float64 `bun:"refunded_amount,type:decimal(15,2),notnull,default:0" json:"refunded_amount"`     // Total refunded
	RefundableAmount float64 `bun:"refundable_amount,type:decimal(15,2),notnull,default:0" json:"refundable_amount"` // Available for refund

	// Payment Information
	PaymentMethod    string       `bun:"payment_method" json:"payment_method"`        // e.g., "credit_card", "bank_transfer", "e_wallet"
	PaymentStatus    string       `bun:"payment_status" json:"payment_status"`        // "pending", "authorized", "captured", "failed", "refunded"
	PaymentProvider  string       `bun:"payment_provider" json:"payment_provider"`    // e.g., "stripe", "midtrans", "xendit"
	TransactionID    string       `bun:"transaction_id,unique" json:"transaction_id"` // External payment provider transaction ID
	PaymentReference string       `bun:"payment_reference" json:"payment_reference"`  // Additional payment reference
	PaidAt           bun.NullTime `bun:"paid_at" json:"paid_at"`                      // When payment was completed

	// Shipping Information
	ShippingMethod    string       `bun:"shipping_method" json:"shipping_method"`       // e.g., "standard", "express", "same_day"
	ShippingProvider  string       `bun:"shipping_provider" json:"shipping_provider"`   // e.g., "jne", "jnt", "sicepat"
	TrackingNumber    string       `bun:"tracking_number" json:"tracking_number"`       // Shipping tracking number
	ShippingStatus    string       `bun:"shipping_status" json:"shipping_status"`       // "pending", "picked_up", "in_transit", "delivered"
	ShippedAt         bun.NullTime `bun:"shipped_at" json:"shipped_at"`                 // When order was shipped
	DeliveredAt       bun.NullTime `bun:"delivered_at" json:"delivered_at"`             // When order was delivered
	EstimatedDelivery bun.NullTime `bun:"estimated_delivery" json:"estimated_delivery"` // Estimated delivery date

	// Shipping Address
	ShippingAddressID   int64  `bun:"shipping_address_id" json:"shipping_address_id"`               // Reference to address table (if normalized)
	ShippingName        string `bun:"shipping_name" json:"shipping_name"`                           // Recipient name
	ShippingPhone       string `bun:"shipping_phone" json:"shipping_phone"`                         // Recipient phone
	ShippingAddressLine string `bun:"shipping_address_line,type:text" json:"shipping_address_line"` // Full address
	ShippingCity        string `bun:"shipping_city" json:"shipping_city"`
	ShippingProvince    string `bun:"shipping_province" json:"shipping_province"`
	ShippingPostalCode  string `bun:"shipping_postal_code" json:"shipping_postal_code"`
	ShippingCountry     string `bun:"shipping_country,default:'ID'" json:"shipping_country"`

	// Additional Information
	Notes              string       `bun:"notes,type:text" json:"notes"`                   // Customer notes
	InternalNotes      string       `bun:"internal_notes,type:text" json:"internal_notes"` // Staff/admin notes
	CancellationReason string       `bun:"cancellation_reason,type:text" json:"cancellation_reason"`
	RefundReason       string       `bun:"refund_reason,type:text" json:"refund_reason"`
	CancelledAt        bun.NullTime `bun:"cancelled_at" json:"cancelled_at"`
	RefundedAt         bun.NullTime `bun:"refunded_at" json:"refunded_at"`

	// Marketing & Analytics
	Source       string `bun:"source" json:"source"`               // e.g., "web", "mobile", "marketplace"
	CampaignCode string `bun:"campaign_code" json:"campaign_code"` // Marketing campaign tracking
	ReferralCode string `bun:"referral_code" json:"referral_code"` // Referral tracking
	UtmSource    string `bun:"utm_source" json:"utm_source"`
	UtmMedium    string `bun:"utm_medium" json:"utm_medium"`
	UtmCampaign  string `bun:"utm_campaign" json:"utm_campaign"`

	// Relationships (loaded separately)
	Items []*OrderItem `bun:"rel:has-many,join:id=order_id" json:"items,omitempty"`
}

// TableName returns the table name for this model.
func (Order) TableName() string {
	return "orders"
}

// BeforeInsert hook to generate order number if not set.
func (o *Order) BeforeInsert(ctx context.Context, query *bun.InsertQuery) error {
	// Call parent hook first
	if err := o.CoreModel.BeforeAppendModel(ctx, query); err != nil {
		return err
	}

	// Generate order number if not set
	if o.OrderNumber == "" {
		o.OrderNumber = generateOrderNumber()
	}

	// Calculate refundable amount
	o.RefundableAmount = o.TotalAmount - o.RefundedAmount

	return nil
}

// BeforeUpdate hook to recalculate amounts.
func (o *Order) BeforeUpdate(ctx context.Context, query *bun.UpdateQuery) error {
	// Call parent hook first
	if err := o.CoreModel.BeforeUpdate(ctx, query); err != nil {
		return err
	}

	// Recalculate refundable amount
	o.RefundableAmount = o.TotalAmount - o.RefundedAmount

	return nil
}

// generateOrderNumber creates a unique order number.
// Format: ORD-YYYYMMDD-NNNN
func generateOrderNumber() string {
	now := time.Now()
	// In production, this should include a counter or UUID to ensure uniqueness
	return fmt.Sprintf("ORD-%s-%04d",
		now.Format("20060102"),
		now.UnixNano()%10000,
	)
}

// IsPaid checks if the order has been paid.
func (o *Order) IsPaid() bool {
	return o.Status == "paid" || o.PaymentStatus == "captured"
}

// CanBeRefunded checks if the order can be refunded.
func (o *Order) CanBeRefunded() bool {
	return o.IsPaid() && o.RefundableAmount > 0 && o.Status != "refunded"
}

// CanBeCancelled checks if the order can be cancelled.
func (o *Order) CanBeCancelled() bool {
	return o.Status == "pending" || o.Status == "payment_pending"
}

// IsDelivered checks if the order has been delivered.
func (o *Order) IsDelivered() bool {
	return o.Status == "delivered" && o.ShippingStatus == "delivered"
}
