package handler

import (
	"context"
	"log"
	"time"

	"consumer-service/internal/events"
)

type OrderHandler interface {
	HandleOrderCreated(ctx context.Context, event events.OrderCreatedEvent) error
}

type InventoryServiceHandler interface {
	DecreaseStock(ctx context.Context, productID string, quantity int) error
}

type orderHandler struct {
	inventoryService InventoryServiceHandler
}

func NewOrderHandler(inventoryService InventoryServiceHandler) OrderHandler {
	return &orderHandler{
		inventoryService: inventoryService,
	}
}

func (h *orderHandler) HandleOrderCreated(ctx context.Context, event events.OrderCreatedEvent) error {
	log.Printf("[ORDER-HANDLER] Processing order %s", event.OrderID)

	// Update inventory for each item in the order
	for _, item := range event.Items {
		log.Printf("[ORDER-HANDLER] Updating stock for product %s, reducing by %d", item.ProductID, item.Quantity)

		// Call inventory service to decrease stock
		if err := h.inventoryService.DecreaseStock(ctx, item.ProductID, item.Quantity); err != nil {
			log.Printf("[ORDER-HANDLER] Failed to decrease stock for product %s: %v", item.ProductID, err)
			return err
		}

		log.Printf("[ORDER-HANDLER] Successfully decreased stock for product %s", item.ProductID)
	}

	log.Printf("[ORDER-HANDLER] Successfully processed order %s at %s", event.OrderID, time.Now().Format(time.RFC3339))
	return nil
}
