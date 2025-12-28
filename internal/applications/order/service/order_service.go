package service

import (
	"context"
	"fmt"
	"ichi-go/internal/applications/order/dto"
	"ichi-go/internal/applications/order/repository"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/logger"
)

type OrderService interface {
	CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
	ProcessPayment(ctx context.Context, orderID string) error
}

type ServiceImpl struct {
	repo     repository.OrderRepository
	producer rabbitmq.MessageProducer
}

func NewOrderService(
	repo repository.OrderRepository,
	producer rabbitmq.MessageProducer,
) *ServiceImpl {
	return &ServiceImpl{
		repo:     repo,
		producer: producer,
	}
}

// ProcessPayment handles payment and publishes events
func (s *ServiceImpl) ProcessPayment(ctx context.Context, orderID string) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	// Simulate payment processing
	paymentSuccess := true // In real: call payment gateway

	if paymentSuccess {
		// Update order status
		order.Status = "paid"
		if err := s.repo.Update(ctx, order); err != nil {
			return err
		}

		// Publish payment completed event
		event := dto.PaymentCompletedEvent{
			EventType:     "payment.completed",
			OrderID:       order.ID,
			UserID:        order.UserID,
			Amount:        order.TotalAmount,
			Currency:      "IDR",
			PaymentMethod: "credit_card",
			TransactionID: "txn_" + order.ID,
		}

		opts := rabbitmq.PublishOptions{}
		if err := s.producer.Publish(ctx, "payment.completed", event, opts); err != nil {
			logger.Errorf("Failed to publish payment event: %v", err)
			// Don't fail the payment - event will be retried or handled separately
		}

		logger.Infof("✅ Payment completed for order %s", orderID)
		return nil
	}

	// Payment failed
	order.Status = "payment_failed"
	if err := s.repo.Update(ctx, order); err != nil {
		return err
	}

	// Publish payment failed event
	event := dto.PaymentFailedEvent{
		EventType: "payment.failed",
		OrderID:   order.ID,
		UserID:    order.UserID,
		Amount:    order.TotalAmount,
		Reason:    "Insufficient funds",
		ErrorCode: "INSUFFICIENT_FUNDS",
	}

	opts := rabbitmq.PublishOptions{}
	if err := s.producer.Publish(ctx, "payment.failed", event, opts); err != nil {
		logger.Errorf("Failed to publish payment failed event: %v", err)
	}

	return fmt.Errorf("payment failed: insufficient funds")
}

// RefundOrder processes refund and publishes event
func (s *ServiceImpl) RefundOrder(ctx context.Context, orderID, reason string) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	if order.Status != "paid" {
		return fmt.Errorf("order not eligible for refund")
	}

	// Process refund with payment gateway
	refundID := "ref_" + order.ID

	// Update order
	order.Status = "refunded"
	if err := s.repo.Update(ctx, order); err != nil {
		return err
	}

	// Publish refund event
	event := dto.PaymentRefundedEvent{
		EventType: "payment.refunded",
		OrderID:   order.ID,
		UserID:    order.UserID,
		Amount:    order.TotalAmount,
		Reason:    reason,
		RefundID:  refundID,
	}

	opts := rabbitmq.PublishOptions{}
	if err := s.producer.Publish(ctx, "payment.refunded", event, opts); err != nil {
		logger.Errorf("Failed to publish refund event: %v", err)
	}

	logger.Infof("✅ Order %s refunded", orderID)
	return nil
}
