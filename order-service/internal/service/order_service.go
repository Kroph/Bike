package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"order-service/internal/cache"
	"order-service/internal/domain"
	"order-service/internal/repository"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error)
	GetOrderByID(ctx context.Context, id string) (domain.Order, error)
	UpdateOrderStatus(ctx context.Context, id string, status domain.OrderStatus) error
	ListOrders(ctx context.Context, filter domain.OrderFilter) ([]domain.Order, int, error)
	GetUserOrders(ctx context.Context, userID string) ([]domain.Order, error)
}

type orderService struct {
	orderRepo   repository.OrderRepository
	natsService NatsService
	cache       cache.Cache
}

func NewOrderService(orderRepo repository.OrderRepository, natsService NatsService, cache cache.Cache) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		natsService: natsService,
		cache:       cache,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	if order.UserID == "" {
		return domain.Order{}, errors.New("user ID is required")
	}

	if len(order.Items) == 0 {
		return domain.Order{}, errors.New("order must contain at least one item")
	}

	var total float64
	for _, item := range order.Items {
		if item.Quantity <= 0 {
			return domain.Order{}, errors.New("item quantity must be greater than zero")
		}
		total += item.Price * float64(item.Quantity)
	}

	order.Total = total
	order.Status = domain.OrderStatusPending

	createdOrder, err := s.orderRepo.Create(ctx, order)
	if err != nil {
		return domain.Order{}, err
	}

	// Publish order created event to NATS
	if err := s.natsService.PublishOrderCreated(createdOrder); err != nil {
		log.Printf("Failed to publish order created event: %v", err)
		// Don't fail the order creation if NATS publish fails
	}

	return createdOrder, nil
}

func (s *orderService) GetOrderByID(ctx context.Context, id string) (domain.Order, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("order:%s", id)
	var cachedOrder domain.Order

	err := s.cache.Get(ctx, cacheKey, &cachedOrder)
	if err == nil {
		log.Printf("Cache hit for order ID: %s", id)
		return cachedOrder, nil
	}

	if err != cache.ErrCacheMiss {
		log.Printf("Cache error for order ID %s: %v", id, err)
	}

	// If not in cache, get from database
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}

	// Store in cache with 10-minute TTL
	if err := s.cache.Set(ctx, cacheKey, order, 10*time.Minute); err != nil {
		log.Printf("Failed to cache order ID %s: %v", id, err)
	}

	return order, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !isValidStatusTransition(order.Status, status) {
		return errors.New("invalid status transition")
	}

	order.Status = status
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("order:%s", id)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		log.Printf("Failed to invalidate cache for order ID %s: %v", id, err)
	}

	return nil
}

func (s *orderService) ListOrders(ctx context.Context, filter domain.OrderFilter) ([]domain.Order, int, error) {
	// Not caching list operations due to complexity of cache invalidation
	return s.orderRepo.List(ctx, filter)
}

func (s *orderService) GetUserOrders(ctx context.Context, userID string) ([]domain.Order, error) {
	// Not caching list operations due to complexity of cache invalidation
	return s.orderRepo.GetUserOrders(ctx, userID)
}

func isValidStatusTransition(current, next domain.OrderStatus) bool {
	switch current {
	case domain.OrderStatusPending:
		return next == domain.OrderStatusPaid || next == domain.OrderStatusCancelled
	case domain.OrderStatusPaid:
		return next == domain.OrderStatusShipped || next == domain.OrderStatusCancelled
	case domain.OrderStatusShipped:
		return next == domain.OrderStatusDelivered || next == domain.OrderStatusCancelled
	case domain.OrderStatusDelivered, domain.OrderStatusCancelled:
		return false
	default:
		return false
	}
}
