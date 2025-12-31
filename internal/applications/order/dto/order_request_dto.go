package dto

type CreateOrderRequest struct {
	UserID      int64          `json:"user_id"`
	Items       []OrderItemDTO `json:"items"`
	TotalAmount float64        `json:"total_amount"`
}

type OrderItemDTO struct {
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}
