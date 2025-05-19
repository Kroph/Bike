package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"order-service/internal/domain"

	"github.com/nats-io/nats.go"
)

type NatsService interface {
	PublishOrderCreated(order domain.Order) error
	Close()
}

type natsService struct {
	conn *nats.Conn
}

func NewNatsService(natsURL string) (NatsService, error) {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %v", err)
	}

	return &natsService{
		conn: conn,
	}, nil
}

func (s *natsService) PublishOrderCreated(order domain.Order) error {
	msg := OrderCreatedEvent{
		OrderID:   order.ID,
		UserID:    order.UserID,
		Total:     order.Total,
		Status:    string(order.Status),
		Items:     make([]OrderItemEvent, len(order.Items)),
		CreatedAt: order.CreatedAt,
	}

	for i, item := range order.Items {
		msg.Items[i] = OrderItemEvent{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal order event: %v", err)
	}

	log.Printf("[NATS-PRODUCER] Publishing order created event for order %s at %s", order.ID, time.Now().Format(time.RFC3339))

	if err := s.conn.Publish("order.created", data); err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	// Flush to ensure the message is sent
	if err := s.conn.Flush(); err != nil {
		return fmt.Errorf("failed to flush message: %v", err)
	}

	log.Printf("[NATS-PRODUCER] Order created event published successfully for order %s", order.ID)
	return nil
}

func (s *natsService) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}

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
