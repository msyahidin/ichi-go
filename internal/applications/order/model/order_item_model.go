package model

import (
	"context"

	"github.com/uptrace/bun"

	"ichi-go/pkg/db/model"
)

// OrderItem represents a line item in an order.
// Each item is a product/service purchased as part of the order.
type OrderItem struct {
	model.CoreModel `bun:"table:order_items,alias:oi"`

	// Relationships
	OrderID int64  `bun:"order_id,notnull" json:"order_id"` // Foreign key to orders table
	Order   *Order `bun:"rel:belongs-to,join:order_id=id" json:"-"`

	// Product Information
	ProductID   int64  `bun:"product_id,notnull" json:"product_id"`     // Reference to products table
	ProductSKU  string `bun:"product_sku,notnull" json:"product_sku"`   // Product SKU for reference
	ProductName string `bun:"product_name,notnull" json:"product_name"` // Snapshot of product name at purchase time
	VariantID   int64  `bun:"variant_id" json:"variant_id"`             // Product variant (e.g., size, color)
	VariantName string `bun:"variant_name" json:"variant_name"`         // Variant description

	// Pricing
	UnitPrice      float64 `bun:"unit_price,type:decimal(15,2),notnull" json:"unit_price"`             // Price per unit at purchase time
	Quantity       int     `bun:"quantity,notnull,default:1" json:"quantity"`                          // Number of units
	SubtotalAmount float64 `bun:"subtotal_amount,type:decimal(15,2),notnull" json:"subtotal_amount"`   // unit_price * quantity
	DiscountAmount float64 `bun:"discount_amount,type:decimal(15,2),default:0" json:"discount_amount"` // Item-level discount
	TaxAmount      float64 `bun:"tax_amount,type:decimal(15,2),default:0" json:"tax_amount"`           // Item-level tax
	TotalAmount    float64 `bun:"total_amount,type:decimal(15,2),notnull" json:"total_amount"`         // Final amount for this item

	// Fulfillment Status
	FulfillmentStatus string `bun:"fulfillment_status,default:'pending'" json:"fulfillment_status"` // "pending", "processing", "shipped", "delivered", "cancelled"

	// Refund Information
	RefundedQuantity int     `bun:"refunded_quantity,default:0" json:"refunded_quantity"`                // How many units have been refunded
	RefundedAmount   float64 `bun:"refunded_amount,type:decimal(15,2),default:0" json:"refunded_amount"` // Total refunded for this item

	// Additional Information
	Notes string `bun:"notes,type:text" json:"notes"` // Item-specific notes (e.g., customization requests)

	// Product Snapshot (optional - for audit trail)
	ProductSnapshot string `bun:"product_snapshot,type:json" json:"product_snapshot,omitempty"` // JSON snapshot of product details at purchase
}

// TableName returns the table name for this model.
func (OrderItem) TableName() string {
	return "order_items"
}

// BeforeInsert calculates amounts before inserting.
func (oi *OrderItem) BeforeInsert(ctx context.Context, query *bun.InsertQuery) error {
	// Call parent hook first
	if err := oi.CoreModel.BeforeAppendModel(ctx, query); err != nil {
		return err
	}

	// Calculate amounts
	oi.SubtotalAmount = oi.UnitPrice * float64(oi.Quantity)
	oi.TotalAmount = oi.SubtotalAmount - oi.DiscountAmount + oi.TaxAmount

	return nil
}

// BeforeUpdate recalculates amounts before updating.
func (oi *OrderItem) BeforeUpdate(ctx context.Context, query *bun.UpdateQuery) error {
	// Call parent hook first
	if err := oi.CoreModel.BeforeUpdate(ctx, query); err != nil {
		return err
	}

	// Recalculate amounts
	oi.SubtotalAmount = oi.UnitPrice * float64(oi.Quantity)
	oi.TotalAmount = oi.SubtotalAmount - oi.DiscountAmount + oi.TaxAmount

	return nil
}

// CanBeRefunded checks if this item can be refunded.
func (oi *OrderItem) CanBeRefunded() bool {
	return oi.RefundedQuantity < oi.Quantity
}

// GetRefundableQuantity returns how many units can still be refunded.
func (oi *OrderItem) GetRefundableQuantity() int {
	return oi.Quantity - oi.RefundedQuantity
}

// GetRefundableAmount returns how much money can still be refunded for this item.
func (oi *OrderItem) GetRefundableAmount() float64 {
	if oi.Quantity == 0 {
		return 0
	}
	// Calculate per-unit amount and multiply by refundable quantity
	perUnitAmount := oi.TotalAmount / float64(oi.Quantity)
	return perUnitAmount * float64(oi.GetRefundableQuantity())
}
