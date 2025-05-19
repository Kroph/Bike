package service

import (
	"context"
	"fmt"
	"log"
	"time"

	inventorypb "proto/inventory"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InventoryService interface {
	DecreaseStock(ctx context.Context, productID string, quantity int) error
	Close()
}

type inventoryService struct {
	conn          *grpc.ClientConn
	productClient inventorypb.ProductServiceClient
}

func NewInventoryService(inventoryURL string) (InventoryService, error) {
	conn, err := grpc.Dial(inventoryURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to inventory service: %v", err)
	}

	productClient := inventorypb.NewProductServiceClient(conn)

	return &inventoryService{
		conn:          conn,
		productClient: productClient,
	}, nil
}

func (s *inventoryService) DecreaseStock(ctx context.Context, productID string, quantity int) error {
	log.Printf("[INVENTORY-SERVICE] Decreasing stock for product %s by %d", productID, quantity)

	// Set timeout for the gRPC call
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Get current product info
	product, err := s.productClient.GetProduct(ctx, &inventorypb.ProductIDRequest{
		Id: productID,
	})
	if err != nil {
		return fmt.Errorf("failed to get product: %v", err)
	}

	// Calculate new stock
	newStock := product.Stock - int32(quantity)
	if newStock < 0 {
		return fmt.Errorf("insufficient stock for product %s", productID)
	}

	// Update product with new stock
	_, err = s.productClient.UpdateProduct(ctx, &inventorypb.UpdateProductRequest{
		Id:          productID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       newStock,
		CategoryId:  product.CategoryId,
	})
	if err != nil {
		return fmt.Errorf("failed to update product stock: %v", err)
	}

	log.Printf("[INVENTORY-SERVICE] Successfully decreased stock for product %s to %d", productID, newStock)
	return nil
}

func (s *inventoryService) Close() {
	if s.conn != nil {
		_ = s.conn.Close()
	}
}
