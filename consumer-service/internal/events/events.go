package events

import "time"

type OrderCreatedEvent struct {
	OrderID    string           `json:"order_id"`
	UserID     string           `json:"user_id"`
	Total      float64          `json:"total"`
	Status     string           `json:"status"`
	Items      []OrderItemEvent `json:"items"`
	PickupDate string           `json:"pickup_date"`
	CreatedAt  time.Time        `json:"created_at"`
}

type OrderItemEvent struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	FrameSize string `json:"frame_size"`
	WheelSize string `json:"wheel_size"`
	Color     string `json:"color"`
	BikeType  string `json:"bike_type"`
}
