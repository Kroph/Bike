package domain

import (
	"time"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

type Payment struct {
	ID        string        `json:"id"`
	OrderID   string        `json:"order_id"`
	Amount    float64       `json:"amount"`
	Status    PaymentStatus `json:"status"`
	Method    string        `json:"method"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}
