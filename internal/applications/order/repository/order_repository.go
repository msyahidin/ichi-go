package repository

import (
	"context"
	"fmt"
	"ichi-go/internal/applications/order/model"
	"ichi-go/pkg/db/query"
	"ichi-go/pkg/db/repository"
	"time"

	"github.com/uptrace/bun"
)

// OrderRepository defines operations for managing orders.
// This interface follows the repository pattern to abstract data access logic.
type OrderRepository interface {
	// Basic CRUD Operations
	Create(ctx context.Context, order *model.Order) (*model.Order, error)
	Update(ctx context.Context, order *model.Order) (*model.Order, error)
	GetByID(ctx context.Context, id string) (*model.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*model.Order, error)
	Delete(ctx context.Context, id string) error

	// Query Operations
	FindByUserID(ctx context.Context, userID int64, scopes ...query.QueryScope) ([]*model.Order, error)
	FindByStatus(ctx context.Context, status string, scopes ...query.QueryScope) ([]*model.Order, error)
	FindByDateRange(ctx context.Context, startDate, endDate time.Time, scopes ...query.QueryScope) ([]*model.Order, error)
	List(ctx context.Context, scopes ...query.QueryScope) ([]*model.Order, error)

	// Pagination
	Paginate(ctx context.Context, page, perPage int, scopes ...query.QueryScope) ([]*model.Order, int, error)

	// Advanced Queries
	FindPendingPayment(ctx context.Context, olderThan time.Duration) ([]*model.Order, error)
	FindRefundable(ctx context.Context) ([]*model.Order, error)
	CountByStatus(ctx context.Context, status string) (int, error)

	// Order Items
	GetOrderWithItems(ctx context.Context, orderID string) (*model.Order, error)
	AddItem(ctx context.Context, item *model.OrderItem) (*model.OrderItem, error)
	UpdateItem(ctx context.Context, item *model.OrderItem) (*model.OrderItem, error)
	RemoveItem(ctx context.Context, itemID int64) error
	GetItemsByOrderID(ctx context.Context, orderID int64) ([]*model.OrderItem, error)

	// Business Logic Helpers
	UpdateStatus(ctx context.Context, orderID string, newStatus string) error
	UpdatePaymentStatus(ctx context.Context, orderID string, paymentStatus string, transactionID string) error
	MarkAsPaid(ctx context.Context, orderID string, transactionID string) error
	MarkAsShipped(ctx context.Context, orderID string, trackingNumber string) error
	MarkAsDelivered(ctx context.Context, orderID string) error
	CancelOrder(ctx context.Context, orderID string, reason string) error
	RefundOrder(ctx context.Context, orderID string, amount float64, reason string) error
}

// OrderRepositoryImpl implements OrderRepository using Bun ORM.
type OrderRepositoryImpl struct {
	*repository.BaseRepository[model.Order]
}

// NewOrderRepository creates a new order repository instance.
func NewOrderRepository(db *bun.DB) OrderRepository {
	baseRepo := repository.NewRepository(db, &model.Order{})
	return &OrderRepositoryImpl{
		BaseRepository: baseRepo,
	}
}

// Create creates a new order with its items in a transaction.
func (r *OrderRepositoryImpl) Create(ctx context.Context, order *model.Order) (*model.Order, error) {
	err := r.DB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Insert order
		if _, err := tx.NewInsert().Model(order).Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert order: %w", err)
		}

		// Insert order items if present
		if len(order.Items) > 0 {
			for _, item := range order.Items {
				item.OrderID = order.ID
			}
			if _, err := tx.NewInsert().Model(&order.Items).Exec(ctx); err != nil {
				return fmt.Errorf("failed to insert order items: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

// Update updates an existing order.
func (r *OrderRepositoryImpl) Update(ctx context.Context, order *model.Order) (*model.Order, error) {
	return r.BaseRepository.Update(ctx, order)
}

// GetByID retrieves an order by its ID.
func (r *OrderRepositoryImpl) GetByID(ctx context.Context, id string) (*model.Order, error) {
	var order model.Order
	err := r.DB().NewSelect().
		Model(&order).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	return &order, nil
}

// GetByOrderNumber retrieves an order by its order number.
func (r *OrderRepositoryImpl) GetByOrderNumber(ctx context.Context, orderNumber string) (*model.Order, error) {
	var order model.Order
	err := r.DB().NewSelect().
		Model(&order).
		Where("order_number = ?", orderNumber).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	return &order, nil
}

// Delete soft deletes an order.
func (r *OrderRepositoryImpl) Delete(ctx context.Context, id string) error {
	_, err := r.DB().NewDelete().
		Model((*model.Order)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// FindByUserID retrieves all orders for a specific user.
func (r *OrderRepositoryImpl) FindByUserID(ctx context.Context, userID int64, scopes ...query.QueryScope) ([]*model.Order, error) {
	userScope := func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("user_id = ?", userID)
	}
	allScopes := append([]query.QueryScope{userScope}, scopes...)
	return r.All(ctx, allScopes...)
}

// FindByStatus retrieves all orders with a specific status.
func (r *OrderRepositoryImpl) FindByStatus(ctx context.Context, status string, scopes ...query.QueryScope) ([]*model.Order, error) {
	statusScope := func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("status = ?", status)
	}
	allScopes := append([]query.QueryScope{statusScope}, scopes...)
	return r.All(ctx, allScopes...)
}

// FindByDateRange retrieves orders within a date range.
func (r *OrderRepositoryImpl) FindByDateRange(ctx context.Context, startDate, endDate time.Time, scopes ...query.QueryScope) ([]*model.Order, error) {
	dateScope := func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("created_at BETWEEN ? AND ?", startDate, endDate)
	}
	allScopes := append([]query.QueryScope{dateScope}, scopes...)
	return r.All(ctx, allScopes...)
}

// List retrieves all orders with optional scopes.
func (r *OrderRepositoryImpl) List(ctx context.Context, scopes ...query.QueryScope) ([]*model.Order, error) {
	return r.All(ctx, scopes...)
}

// Paginate retrieves paginated orders with total count.
func (r *OrderRepositoryImpl) Paginate(ctx context.Context, page, perPage int, scopes ...query.QueryScope) ([]*model.Order, int, error) {
	return r.PaginateWithCount(ctx, page, perPage, scopes...)
}

// FindPendingPayment finds orders with pending payment older than specified duration.
func (r *OrderRepositoryImpl) FindPendingPayment(ctx context.Context, olderThan time.Duration) ([]*model.Order, error) {
	cutoffTime := time.Now().Add(-olderThan)

	var orders []*model.Order
	err := r.DB().NewSelect().
		Model(&orders).
		Where("status = ?", "payment_pending").
		Where("created_at < ?", cutoffTime).
		Order("created_at DESC").
		Scan(ctx)

	return orders, err
}

// FindRefundable finds orders that can be refunded.
func (r *OrderRepositoryImpl) FindRefundable(ctx context.Context) ([]*model.Order, error) {
	var orders []*model.Order
	err := r.DB().NewSelect().
		Model(&orders).
		Where("status = ?", "paid").
		Where("refundable_amount > 0").
		Order("paid_at DESC").
		Scan(ctx)

	return orders, err
}

// CountByStatus counts orders by status.
func (r *OrderRepositoryImpl) CountByStatus(ctx context.Context, status string) (int, error) {
	return r.DB().NewSelect().
		Model((*model.Order)(nil)).
		Where("status = ?", status).
		Count(ctx)
}

// GetOrderWithItems retrieves an order with all its items.
func (r *OrderRepositoryImpl) GetOrderWithItems(ctx context.Context, orderID string) (*model.Order, error) {
	var order model.Order
	err := r.DB().NewSelect().
		Model(&order).
		Relation("Items").
		Where("o.id = ?", orderID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	return &order, nil
}

// AddItem adds an item to an order.
func (r *OrderRepositoryImpl) AddItem(ctx context.Context, item *model.OrderItem) (*model.OrderItem, error) {
	_, err := r.DB().NewInsert().
		Model(item).
		Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to add order item: %w", err)
	}

	return item, nil
}

// UpdateItem updates an order item.
func (r *OrderRepositoryImpl) UpdateItem(ctx context.Context, item *model.OrderItem) (*model.OrderItem, error) {
	_, err := r.DB().NewUpdate().
		Model(item).
		WherePK().
		Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to update order item: %w", err)
	}

	return item, nil
}

// RemoveItem removes an item from an order.
func (r *OrderRepositoryImpl) RemoveItem(ctx context.Context, itemID int64) error {
	_, err := r.DB().NewDelete().
		Model((*model.OrderItem)(nil)).
		Where("id = ?", itemID).
		Exec(ctx)

	return err
}

// GetItemsByOrderID retrieves all items for an order.
func (r *OrderRepositoryImpl) GetItemsByOrderID(ctx context.Context, orderID int64) ([]*model.OrderItem, error) {
	var items []*model.OrderItem
	err := r.DB().NewSelect().
		Model(&items).
		Where("order_id = ?", orderID).
		Scan(ctx)

	return items, err
}

// UpdateStatus updates order status.
func (r *OrderRepositoryImpl) UpdateStatus(ctx context.Context, orderID string, newStatus string) error {
	_, err := r.DB().NewUpdate().
		Model((*model.Order)(nil)).
		Set("status = ?", newStatus).
		Where("id = ?", orderID).
		Exec(ctx)

	return err
}

// UpdatePaymentStatus updates payment status and transaction ID.
func (r *OrderRepositoryImpl) UpdatePaymentStatus(ctx context.Context, orderID string, paymentStatus string, transactionID string) error {
	_, err := r.DB().NewUpdate().
		Model((*model.Order)(nil)).
		Set("payment_status = ?", paymentStatus).
		Set("transaction_id = ?", transactionID).
		Where("id = ?", orderID).
		Exec(ctx)

	return err
}

// MarkAsPaid marks an order as paid.
func (r *OrderRepositoryImpl) MarkAsPaid(ctx context.Context, orderID string, transactionID string) error {
	now := time.Now()
	_, err := r.DB().NewUpdate().
		Model((*model.Order)(nil)).
		Set("status = ?", "paid").
		Set("payment_status = ?", "captured").
		Set("transaction_id = ?", transactionID).
		Set("paid_at = ?", now).
		Where("id = ?", orderID).
		Exec(ctx)

	return err
}

// MarkAsShipped marks an order as shipped.
func (r *OrderRepositoryImpl) MarkAsShipped(ctx context.Context, orderID string, trackingNumber string) error {
	now := time.Now()
	_, err := r.DB().NewUpdate().
		Model((*model.Order)(nil)).
		Set("status = ?", "shipped").
		Set("shipping_status = ?", "in_transit").
		Set("tracking_number = ?", trackingNumber).
		Set("shipped_at = ?", now).
		Where("id = ?", orderID).
		Exec(ctx)

	return err
}

// MarkAsDelivered marks an order as delivered.
func (r *OrderRepositoryImpl) MarkAsDelivered(ctx context.Context, orderID string) error {
	now := time.Now()
	_, err := r.DB().NewUpdate().
		Model((*model.Order)(nil)).
		Set("status = ?", "delivered").
		Set("shipping_status = ?", "delivered").
		Set("delivered_at = ?", now).
		Where("id = ?", orderID).
		Exec(ctx)

	return err
}

// CancelOrder cancels an order with a reason.
func (r *OrderRepositoryImpl) CancelOrder(ctx context.Context, orderID string, reason string) error {
	now := time.Now()
	_, err := r.DB().NewUpdate().
		Model((*model.Order)(nil)).
		Set("status = ?", "cancelled").
		Set("cancellation_reason = ?", reason).
		Set("cancelled_at = ?", now).
		Where("id = ?", orderID).
		Exec(ctx)

	return err
}

// RefundOrder processes a refund for an order.
func (r *OrderRepositoryImpl) RefundOrder(ctx context.Context, orderID string, amount float64, reason string) error {
	now := time.Now()

	return r.DB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Get current order
		var order model.Order
		err := tx.NewSelect().
			Model(&order).
			Where("id = ?", orderID).
			For("UPDATE").
			Scan(ctx)

		if err != nil {
			return fmt.Errorf("order not found: %w", err)
		}

		// Validate refund amount
		if amount > order.RefundableAmount {
			return fmt.Errorf("refund amount exceeds refundable amount")
		}

		// Calculate new refunded amount
		newRefundedAmount := order.RefundedAmount + amount
		newRefundableAmount := order.TotalAmount - newRefundedAmount

		// Determine new status
		newStatus := order.Status
		if newRefundableAmount == 0 {
			newStatus = "refunded"
		}

		// Update order
		_, err = tx.NewUpdate().
			Model((*model.Order)(nil)).
			Set("status = ?", newStatus).
			Set("refunded_amount = ?", newRefundedAmount).
			Set("refundable_amount = ?", newRefundableAmount).
			Set("refund_reason = ?", reason).
			Set("refunded_at = ?", now).
			Where("id = ?", orderID).
			Exec(ctx)

		return err
	})
}
