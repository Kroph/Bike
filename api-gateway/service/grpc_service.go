package service

import (
	"context"
	"fmt"
	"time"

	inventorypb "proto/inventory"
	orderpb "proto/order"
	userpb "proto/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClients struct {
	userClient      userpb.UserServiceClient
	inventoryClient struct {
		product  inventorypb.ProductServiceClient
		category inventorypb.CategoryServiceClient
	}
	orderClient struct {
		order   orderpb.OrderServiceClient
		payment orderpb.PaymentServiceClient
	}
}

func NewGrpcClients(userServiceURL, inventoryServiceURL, orderServiceURL string) (*GrpcClients, error) {
	clients := &GrpcClients{}

	// Set up connection to User Service
	userConn, err := grpc.Dial(userServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	clients.userClient = userpb.NewUserServiceClient(userConn)

	// Set up connection to Inventory Service
	inventoryConn, err := grpc.Dial(inventoryServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	clients.inventoryClient.product = inventorypb.NewProductServiceClient(inventoryConn)
	clients.inventoryClient.category = inventorypb.NewCategoryServiceClient(inventoryConn)

	// Set up connection to Order Service
	orderConn, err := grpc.Dial(orderServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	clients.orderClient.order = orderpb.NewOrderServiceClient(orderConn)
	clients.orderClient.payment = orderpb.NewPaymentServiceClient(orderConn)

	return clients, nil
}

// User Service methods

func (c *GrpcClients) RegisterUser(ctx context.Context, username, email, password string) (*userpb.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.userClient.RegisterUser(ctx, &userpb.RegisterUserRequest{
		Username: username,
		Email:    email,
		Password: password,
		Role:     userpb.UserRole_USER,
	})
}

func (c *GrpcClients) AuthenticateUser(ctx context.Context, email, password string) (*userpb.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.userClient.AuthenticateUser(ctx, &userpb.AuthRequest{
		Email:    email,
		Password: password,
	})
}

func (c *GrpcClients) GetUserWithRole(ctx context.Context, userID string) (*userpb.UserProfile, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.userClient.GetUserProfile(ctx, &userpb.UserIDRequest{
		UserId: userID,
	})
}

func (c *GrpcClients) ListOrders(ctx context.Context, filter *orderpb.OrderFilter) (*orderpb.ListOrdersResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.orderClient.order.ListOrders(ctx, &orderpb.ListOrdersRequest{
		Filter: filter,
	})
}

func (c *GrpcClients) GetUserProfile(ctx context.Context, userID string) (*userpb.UserProfile, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.userClient.GetUserProfile(ctx, &userpb.UserIDRequest{
		UserId: userID,
	})
}

// Inventory Service - Product methods

func (c *GrpcClients) CreateProduct(ctx context.Context, req *inventorypb.CreateProductRequest) (*inventorypb.ProductResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.inventoryClient.product.CreateProduct(ctx, req)
}

func (c *GrpcClients) UpdateProduct(ctx context.Context, req *inventorypb.UpdateProductRequest) (*inventorypb.ProductResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.inventoryClient.product.UpdateProduct(ctx, req)
}

func (c *GrpcClients) ListProducts(ctx context.Context, filter *inventorypb.ProductFilter) (*inventorypb.ListProductsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.inventoryClient.product.ListProducts(ctx, &inventorypb.ListProductsRequest{
		Filter: filter,
	})
}

func (c *GrpcClients) GetProduct(ctx context.Context, productID string) (*inventorypb.ProductResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.inventoryClient.product.GetProduct(ctx, &inventorypb.ProductIDRequest{
		Id: productID,
	})
}

func (c *GrpcClients) DeleteProduct(ctx context.Context, productID string) (*inventorypb.DeleteResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.inventoryClient.product.DeleteProduct(ctx, &inventorypb.ProductIDRequest{
		Id: productID,
	})
}

func (c *GrpcClients) CheckStock(ctx context.Context, items []*inventorypb.ProductQuantity) (*inventorypb.CheckStockResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.inventoryClient.product.CheckStock(ctx, &inventorypb.CheckStockRequest{
		Items: items,
	})
}

// Inventory Service - Category methods
func (c *GrpcClients) CreateCategory(ctx context.Context, req *inventorypb.CreateCategoryRequest) (*inventorypb.CategoryResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.inventoryClient.category.CreateCategory(ctx, req)
}

func (c *GrpcClients) GetCategory(ctx context.Context, categoryID string) (*inventorypb.CategoryResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.inventoryClient.category.GetCategory(ctx, &inventorypb.CategoryIDRequest{
		Id: categoryID,
	})
}

func (c *GrpcClients) UpdateCategory(ctx context.Context, req *inventorypb.UpdateCategoryRequest) (*inventorypb.CategoryResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.inventoryClient.category.UpdateCategory(ctx, req)
}

func (c *GrpcClients) DeleteCategory(ctx context.Context, categoryID string) (*inventorypb.DeleteCategoryResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.inventoryClient.category.DeleteCategory(ctx, &inventorypb.CategoryIDRequest{
		Id: categoryID,
	})
}

func (c *GrpcClients) ListCategories(ctx context.Context) (*inventorypb.ListCategoriesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.inventoryClient.category.ListCategories(ctx, &inventorypb.ListCategoriesRequest{})
}

// Order Service - Order methods
func (c *GrpcClients) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.OrderResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.orderClient.order.CreateOrder(ctx, req)
}

func (c *GrpcClients) GetOrder(ctx context.Context, orderID string) (*orderpb.OrderResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.orderClient.order.GetOrder(ctx, &orderpb.OrderIDRequest{
		Id: orderID,
	})
}

func (c *GrpcClients) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.OrderResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.orderClient.order.UpdateOrderStatus(ctx, req)
}

func (c *GrpcClients) GetUserOrders(ctx context.Context, userID string) (*orderpb.ListOrdersResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.orderClient.order.GetUserOrders(ctx, &orderpb.UserIDRequest{
		UserId: userID,
	})
}

func (c *GrpcClients) GenerateVerificationCode(ctx context.Context, userID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response, err := c.userClient.GenerateVerificationCode(ctx, &userpb.GenerateCodeRequest{
		UserId: userID,
	})
	if err != nil {
		return "", err
	}

	if !response.Success {
		return "", fmt.Errorf("failed to generate verification code: %s", response.Message)
	}

	return response.Code, nil
}

func (c *GrpcClients) VerifyEmailCode(ctx context.Context, userID, code string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response, err := c.userClient.VerifyEmail(ctx, &userpb.VerifyEmailRequest{
		UserId: userID,
		Code:   code,
	})
	if err != nil {
		return false, err
	}

	return response.Success, nil
}
