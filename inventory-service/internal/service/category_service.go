package service

import (
	"context"

	"inventory-service/internal/domain"
	"inventory-service/internal/repository"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, category domain.Category) (domain.Category, error)
	GetCategoryByID(ctx context.Context, id string) (domain.Category, error)
	UpdateCategory(ctx context.Context, category domain.Category) error
	DeleteCategory(ctx context.Context, id string) error
	ListCategories(ctx context.Context) ([]domain.Category, error)
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *categoryService) CreateCategory(ctx context.Context, category domain.Category) (domain.Category, error) {
	return s.categoryRepo.Create(ctx, category)
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id string) (domain.Category, error) {
	return s.categoryRepo.GetByID(ctx, id)
}

func (s *categoryService) UpdateCategory(ctx context.Context, category domain.Category) error {
	return s.categoryRepo.Update(ctx, category)
}

func (s *categoryService) DeleteCategory(ctx context.Context, id string) error {
	return s.categoryRepo.Delete(ctx, id)
}

func (s *categoryService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.categoryRepo.List(ctx)
}
