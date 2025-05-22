package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"inventory-service/internal/domain"

	"github.com/google/uuid"
)

type CategoryRepository interface {
	Create(ctx context.Context, category domain.Category) (domain.Category, error)
	GetByID(ctx context.Context, id string) (domain.Category, error)
	Update(ctx context.Context, category domain.Category) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]domain.Category, error)
	// Add missing methods
	ExistsByID(ctx context.Context, id string) (bool, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	GetByName(ctx context.Context, name string) (domain.Category, error)
	GetCategoriesWithProductCount(ctx context.Context) ([]CategoryWithCount, error)
}

type CategoryWithCount struct {
	Category     domain.Category
	ProductCount int
}

type PostgresCategoryRepository struct {
	db *sql.DB
}

func NewPostgresCategoryRepository(db *sql.DB) CategoryRepository {
	return &PostgresCategoryRepository{
		db: db,
	}
}

func (r *PostgresCategoryRepository) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
	// Validate input
	if err := r.validateCategory(category); err != nil {
		return domain.Category{}, err
	}

	// Check if category with same name already exists
	exists, err := r.ExistsByName(ctx, category.Name)
	if err != nil {
		return domain.Category{}, errors.New("failed to check category name uniqueness")
	}
	if exists {
		return domain.Category{}, errors.New("category with this name already exists")
	}

	category.ID = uuid.New().String()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	query := `
		INSERT INTO categories (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, created_at, updated_at`

	err = r.db.QueryRowContext(
		ctx,
		query,
		category.ID,
		strings.TrimSpace(category.Name),
		strings.TrimSpace(category.Description),
		category.CreatedAt,
		category.UpdatedAt,
	).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.Category{}, errors.New("category with this name already exists")
		}
		return domain.Category{}, errors.New("failed to create category")
	}

	return category, nil
}

func (r *PostgresCategoryRepository) GetByID(ctx context.Context, id string) (domain.Category, error) {
	if id == "" {
		return domain.Category{}, errors.New("category ID is required")
	}

	query := `
		SELECT id, name, description, created_at, updated_at
		FROM categories
		WHERE id = $1`

	var category domain.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Category{}, errors.New("category not found")
		}
		return domain.Category{}, errors.New("failed to get category")
	}

	return category, nil
}

func (r *PostgresCategoryRepository) Update(ctx context.Context, category domain.Category) error {
	if category.ID == "" {
		return errors.New("category ID is required")
	}

	if err := r.validateCategory(category); err != nil {
		return err
	}

	// Check if another category with same name exists (excluding current category)
	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE LOWER(name) = LOWER($1) AND id != $2)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(category.Name), category.ID).Scan(&exists)
	if err != nil {
		return errors.New("failed to check category name uniqueness")
	}
	if exists {
		return errors.New("another category with this name already exists")
	}

	category.UpdatedAt = time.Now()

	updateQuery := `
		UPDATE categories
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4`

	result, err := r.db.ExecContext(
		ctx,
		updateQuery,
		strings.TrimSpace(category.Name),
		strings.TrimSpace(category.Description),
		category.UpdatedAt,
		category.ID,
	)

	if err != nil {
		if isDuplicateKeyError(err) {
			return errors.New("category with this name already exists")
		}
		return errors.New("failed to update category")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("failed to check update result")
	}

	if rowsAffected == 0 {
		return errors.New("category not found")
	}

	return nil
}

func (r *PostgresCategoryRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("category ID is required")
	}

	// Check if category exists
	exists, err := r.ExistsByID(ctx, id)
	if err != nil {
		return errors.New("failed to check if category exists")
	}
	if !exists {
		return errors.New("category not found")
	}

	// Check if category has products
	hasProducts, err := r.hasProducts(ctx, id)
	if err != nil {
		return errors.New("failed to check category products")
	}

	if hasProducts {
		return errors.New("cannot delete category with existing products")
	}

	query := `DELETE FROM categories WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.New("failed to delete category")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("failed to check delete result")
	}

	if rowsAffected == 0 {
		return errors.New("category not found")
	}

	return nil
}

func (r *PostgresCategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM categories
		WHERE name IS NOT NULL AND name != ''
		ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.New("failed to list categories")
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var category domain.Category
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, errors.New("failed to scan category")
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("error reading category rows")
	}

	return categories, nil
}

func (r *PostgresCategoryRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, errors.New("category ID is required")
	}

	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, errors.New("failed to check if category exists")
	}

	return exists, nil
}

func (r *PostgresCategoryRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	if strings.TrimSpace(name) == "" {
		return false, errors.New("category name is required")
	}

	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE LOWER(name) = LOWER($1))`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(name)).Scan(&exists)
	if err != nil {
		return false, errors.New("failed to check if category exists")
	}

	return exists, nil
}

func (r *PostgresCategoryRepository) GetByName(ctx context.Context, name string) (domain.Category, error) {
	if strings.TrimSpace(name) == "" {
		return domain.Category{}, errors.New("category name is required")
	}

	query := `
		SELECT id, name, description, created_at, updated_at
		FROM categories
		WHERE LOWER(name) = LOWER($1)`

	var category domain.Category
	err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(name)).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Category{}, errors.New("category not found")
		}
		return domain.Category{}, errors.New("failed to get category")
	}

	return category, nil
}

func (r *PostgresCategoryRepository) GetCategoriesWithProductCount(ctx context.Context) ([]CategoryWithCount, error) {
	query := `
		SELECT 
			c.id, c.name, c.description, c.created_at, c.updated_at,
			COALESCE(COUNT(p.id), 0) as product_count
		FROM categories c
		LEFT JOIN products p ON c.id = p.category_id
		GROUP BY c.id, c.name, c.description, c.created_at, c.updated_at
		ORDER BY c.name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.New("failed to get categories with product count")
	}
	defer rows.Close()

	var results []CategoryWithCount
	for rows.Next() {
		var result CategoryWithCount
		err := rows.Scan(
			&result.Category.ID,
			&result.Category.Name,
			&result.Category.Description,
			&result.Category.CreatedAt,
			&result.Category.UpdatedAt,
			&result.ProductCount,
		)
		if err != nil {
			return nil, errors.New("failed to scan category with count")
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("error reading category rows")
	}

	return results, nil
}

// Helper methods

func (r *PostgresCategoryRepository) hasProducts(ctx context.Context, categoryID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE category_id = $1)`

	var hasProducts bool
	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(&hasProducts)
	if err != nil {
		return false, err
	}

	return hasProducts, nil
}

func (r *PostgresCategoryRepository) validateCategory(category domain.Category) error {
	name := strings.TrimSpace(category.Name)
	if name == "" {
		return errors.New("category name is required")
	}

	if len(name) > 100 {
		return errors.New("category name is too long (max 100 characters)")
	}

	if len(name) < 2 {
		return errors.New("category name is too short (min 2 characters)")
	}

	if len(category.Description) > 1000 {
		return errors.New("category description is too long (max 1000 characters)")
	}

	// Check for invalid characters that might cause issues
	if strings.ContainsAny(name, "<>\"'&;--") {
		return errors.New("category name contains invalid characters")
	}

	// Check for SQL injection patterns
	lowerName := strings.ToLower(name)
	sqlKeywords := []string{"select", "insert", "update", "delete", "drop", "union", "script"}
	for _, keyword := range sqlKeywords {
		if strings.Contains(lowerName, keyword) {
			return errors.New("category name contains prohibited keywords")
		}
	}

	return nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate key value violates unique constraint") ||
		strings.Contains(errStr, "unique constraint failed") ||
		strings.Contains(errStr, "violates unique constraint") ||
		strings.Contains(errStr, "duplicate entry") ||
		strings.Contains(errStr, "unique violation")
}
