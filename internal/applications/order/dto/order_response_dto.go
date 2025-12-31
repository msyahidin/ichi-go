package dto

type OrderResponse struct {
	ID          string  `json:"id"`
	UserID      int64   `json:"user_id"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
}
