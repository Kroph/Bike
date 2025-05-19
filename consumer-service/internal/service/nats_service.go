package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"consumer-service/internal/events"

	"github.com/nats-io/nats.go"
)

type OrderEventHandler interface {
	HandleOrderCreated(ctx context.Context, event events.OrderCreatedEvent) error
}

type NatsService interface {
	StartConsuming(ctx context.Context) error
	Close()
}

type natsService struct {
	conn         *nats.Conn
	subscription *nats.Subscription
	handler      OrderEventHandler
}

func NewNatsService(natsURL string, orderHandler OrderEventHandler) (NatsService, error) {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %v", err)
	}

	return &natsService{
		conn:    conn,
		handler: orderHandler,
	}, nil
}

func (s *natsService) StartConsuming(ctx context.Context) error {
	sub, err := s.conn.Subscribe("bicycle.order.created", func(msg *nats.Msg) {
		log.Printf("[NATS-CONSUMER] Received bicycle order from subject %s at %s", msg.Subject, time.Now().Format(time.RFC3339))

		var orderEvent events.OrderCreatedEvent
		if err := json.Unmarshal(msg.Data, &orderEvent); err != nil {
			log.Printf("[NATS-CONSUMER] Failed to unmarshal bicycle order event: %v", err)
			return
		}

		log.Printf("[NATS-CONSUMER] Processing bicycle order %s created at %s",
			orderEvent.OrderID, orderEvent.CreatedAt.Format(time.RFC3339))

		if err := s.handler.HandleOrderCreated(ctx, orderEvent); err != nil {
			log.Printf("[NATS-CONSUMER] Failed to handle bicycle order created event: %v", err)
			return
		}

		log.Printf("[NATS-CONSUMER] Successfully processed bicycle order %s", orderEvent.OrderID)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to bicycle.order.created: %v", err)
	}

	s.subscription = sub
	return nil
}

func (s *natsService) Close() {
	if s.subscription != nil {
		_ = s.subscription.Unsubscribe()
	}
	if s.conn != nil {
		s.conn.Close()
	}
}
