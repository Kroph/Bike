package domain

import (
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Status    OrderStatus `json:"status"`
	Total     float64     `json:"total"`
	Items     []OrderItem `json:"items"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        string  `json:"id"`
	OrderID   string  `json:"order_id"`
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

type OrderFilter struct {
	UserID   string
	Status   OrderStatus
	FromDate *time.Time
	ToDate   *time.Time
	Page     int
	PageSize int
}
