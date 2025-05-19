package events

import "time"

type OrderCreatedEvent struct {
	OrderID   string           `json:"order_id"`
	UserID    string           `json:"user_id"`
	Total     float64          `json:"total"`
	Status    string           `json:"status"`
	Items     []OrderItemEvent `json:"items"`
	CreatedAt time.Time        `json:"created_at"`
}

type OrderItemEvent struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}
