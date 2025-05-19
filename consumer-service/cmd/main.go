package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"consumer-service/config"
	"consumer-service/internal/handler"
	"consumer-service/internal/service"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize inventory service
	inventoryService, err := service.NewInventoryService(cfg.Services.Inventory.GrpcURL)
	if err != nil {
		log.Fatalf("Failed to initialize inventory service: %v", err)
	}
	defer inventoryService.Close()

	// Initialize order handler - cast inventoryService to the interface expected by handler
	orderHandler := handler.NewOrderHandler(inventoryService)

	// Initialize NATS service - cast orderHandler to the interface expected by nats service
	natsService, err := service.NewNatsService(cfg.NATS.URL, orderHandler)
	if err != nil {
		log.Fatalf("Failed to initialize NATS service: %v", err)
	}
	defer natsService.Close()

	// Start consuming messages
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := natsService.StartConsuming(ctx); err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	log.Println("Consumer service started.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down consumer service...")
	cancel()
	log.Println("Consumer service stopped.")
}
