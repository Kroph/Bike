package service_test

import (
	"context"
	"testing"
	"time"

	"api-gateway/service"
	inventorypb "proto/inventory"

	"github.com/stretchr/testify/require"
)

func TestCreateCategoryAndProduct(t *testing.T) {
	grpcClients, err := service.NewGrpcClients("localhost:50051", "localhost:50052", "localhost:50053")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	categoryResp, err := grpcClients.CreateCategory(ctx, &inventorypb.CreateCategoryRequest{
		Name:        "Test Category",
		Description: "Category for testing",
	})
	require.NoError(t, err)
	require.NotEmpty(t, categoryResp.Id)

	productResp, err := grpcClients.CreateProduct(ctx, &inventorypb.CreateProductRequest{
		Name:        "Test Bike",
		Description: "Testing bike creation",
		Price:       999.99,
		Stock:       5,
		CategoryId:  categoryResp.Id,
	})
	require.NoError(t, err)
	require.Equal(t, "Test Bike", productResp.Name)
	require.Equal(t, categoryResp.Id, productResp.CategoryId)
}

func TestCheckStockAvailable(t *testing.T) {
	grpcClients, err := service.NewGrpcClients("localhost:50051", "localhost:50052", "localhost:50053")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	items := []*inventorypb.ProductQuantity{
		{
			ProductId: "existing-product-id",
			Quantity:  1,
		},
	}

	resp, err := grpcClients.CheckStock(ctx, items)
	require.NoError(t, err)
	require.True(t, resp.Available)
}
