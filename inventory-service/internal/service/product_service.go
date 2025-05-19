package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"inventory-service/internal/cache"
	"inventory-service/internal/domain"
	"inventory-service/internal/repository"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product domain.Product) (domain.Product, error)
	GetProductByID(ctx context.Context, id string) (domain.Product, error)
	UpdateProduct(ctx context.Context, product domain.Product) error
	DeleteProduct(ctx context.Context, id string) error
	ListProducts(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, int, error)
}

type productService struct {
	productRepo repository.ProductRepository
	cache       cache.Cache
}

func NewProductService(productRepo repository.ProductRepository, cache cache.Cache) ProductService {
	return &productService{
		productRepo: productRepo,
		cache:       cache,
	}
}

func (s *productService) CreateProduct(ctx context.Context, product domain.Product) (domain.Product, error) {
	return s.productRepo.Create(ctx, product)
}

func (s *productService) GetProductByID(ctx context.Context, id string) (domain.Product, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("product:%s", id)
	var cachedProduct domain.Product

	err := s.cache.Get(ctx, cacheKey, &cachedProduct)
	if err == nil {
		log.Printf("Cache hit for product ID: %s", id)
		return cachedProduct, nil
	}

	if err != cache.ErrCacheMiss {
		log.Printf("Cache error for product ID %s: %v", id, err)
	}

	// If not in cache, get from database
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return domain.Product{}, err
	}

	if err := s.cache.Set(ctx, cacheKey, product, 5*time.Minute); err != nil {
		log.Printf("Failed to cache product ID %s: %v", id, err)
	}

	return product, nil
}

func (s *productService) UpdateProduct(ctx context.Context, product domain.Product) error {
	if err := s.productRepo.Update(ctx, product); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("product:%s", product.ID)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		log.Printf("Failed to invalidate cache for product ID %s: %v", product.ID, err)
	}

	return nil
}

func (s *productService) DeleteProduct(ctx context.Context, id string) error {
	if err := s.productRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("product:%s", id)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		log.Printf("Failed to invalidate cache for product ID %s: %v", id, err)
	}

	return nil
}

func (s *productService) ListProducts(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, int, error) {
	// Not caching list operations due to complexity of cache invalidation
	return s.productRepo.List(ctx, filter)
}
