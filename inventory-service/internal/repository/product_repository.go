package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"inventory-service/internal/domain"

	"github.com/google/uuid"
)

type ProductRepository interface {
	Create(ctx context.Context, product domain.Product) (domain.Product, error)
	GetByID(ctx context.Context, id string) (domain.Product, error)
	Update(ctx context.Context, product domain.Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, int, error)
	ExistsByID(ctx context.Context, id string) (bool, error)
	GetByCategory(ctx context.Context, categoryID string) ([]domain.Product, error)
	UpdateStock(ctx context.Context, productID string, newStock int) error
}

type PostgresProductRepository struct {
	db *sql.DB
}

func NewPostgresProductRepository(db *sql.DB) ProductRepository {
	return &PostgresProductRepository{
		db: db,
	}
}

func (r *PostgresProductRepository) Create(ctx context.Context, product domain.Product) (domain.Product, error) {
	// Validate input
	if err := r.validateProduct(product); err != nil {
		return domain.Product{}, err
	}

	product.ID = uuid.New().String()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	query := `
		INSERT INTO products (id, name, description, price, stock, category_id, 
		                      frame_size, wheel_size, color, weight, bike_type, 
		                      created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, name, description, price, stock, category_id, 
		          frame_size, wheel_size, color, weight, bike_type, 
		          created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx,
		query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.CategoryID,
		nullString(product.FrameSize),
		nullString(product.WheelSize),
		nullString(product.Color),
		nullFloat64(product.Weight),
		nullString(product.BikeType),
		product.CreatedAt,
		product.UpdatedAt,
	).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.CategoryID,
		&product.FrameSize,
		&product.WheelSize,
		&product.Color,
		&product.Weight,
		&product.BikeType,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if isForeignKeyError(err) {
			return domain.Product{}, errors.New("invalid category ID")
		}
		return domain.Product{}, errors.New("failed to create product")
	}

	return product, nil
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id string) (domain.Product, error) {
	if id == "" {
		return domain.Product{}, errors.New("product ID is required")
	}

	query := `
        SELECT id, name, description, price, stock, category_id, 
               COALESCE(frame_size, '') as frame_size,
               COALESCE(wheel_size, '') as wheel_size,
               COALESCE(color, '') as color,
               COALESCE(weight, 0) as weight,
               COALESCE(bike_type, '') as bike_type,
               created_at, updated_at
        FROM products
        WHERE id = $1`

	var product domain.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.CategoryID,
		&product.FrameSize,
		&product.WheelSize,
		&product.Color,
		&product.Weight,
		&product.BikeType,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Product{}, errors.New("product not found")
		}
		return domain.Product{}, errors.New("failed to get product")
	}

	return product, nil
}

func (r *PostgresProductRepository) Update(ctx context.Context, product domain.Product) error {
	if product.ID == "" {
		return errors.New("product ID is required")
	}

	if err := r.validateProduct(product); err != nil {
		return err
	}

	product.UpdatedAt = time.Now()

	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, stock = $4, category_id = $5,
		    frame_size = $6, wheel_size = $7, color = $8, weight = $9, bike_type = $10, 
		    updated_at = $11
		WHERE id = $12`

	result, err := r.db.ExecContext(
		ctx,
		query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.CategoryID,
		nullString(product.FrameSize),
		nullString(product.WheelSize),
		nullString(product.Color),
		nullFloat64(product.Weight),
		nullString(product.BikeType),
		product.UpdatedAt,
		product.ID,
	)

	if err != nil {
		if isForeignKeyError(err) {
			return errors.New("invalid category ID")
		}
		return errors.New("failed to update product")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("failed to check update result")
	}

	if rowsAffected == 0 {
		return errors.New("product not found")
	}

	return nil
}

func (r *PostgresProductRepository) List(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, int, error) {
	// Build the query dynamically based on filters
	baseQuery := `
        SELECT id, name, description, price, stock, category_id, 
               COALESCE(frame_size, '') as frame_size,
               COALESCE(wheel_size, '') as wheel_size,
               COALESCE(color, '') as color,
               COALESCE(weight, 0) as weight,
               COALESCE(bike_type, '') as bike_type,
               created_at, updated_at
        FROM products`

	countQuery := `SELECT COUNT(*) FROM products`

	whereClause, args := r.buildWhereClause(filter)

	if whereClause != "" {
		baseQuery += " WHERE " + whereClause
		countQuery += " WHERE " + whereClause
	}

	// Add pagination
	limit := 10
	offset := 0

	if filter.PageSize > 0 {
		limit = filter.PageSize
	}

	if filter.Page > 0 {
		offset = (filter.Page - 1) * limit
	}

	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	// Execute the main query
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, errors.New("failed to list products")
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.CategoryID,
			&product.FrameSize,
			&product.WheelSize,
			&product.Color,
			&product.Weight,
			&product.BikeType,
			&product.CreatedAt,
			&product.UpdatedAt,
		)

		if err != nil {
			return nil, 0, errors.New("failed to scan product")
		}

		products = append(products, product)
	}

	// Get the total count
	var total int
	countArgs := args[:len(args)-2] // Remove limit and offset
	err = r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, errors.New("failed to get product count")
	}

	return products, total, nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("product ID is required")
	}

	query := `DELETE FROM products WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.New("failed to delete product")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("failed to check delete result")
	}

	if rowsAffected == 0 {
		return errors.New("product not found")
	}

	return nil
}

func (r *PostgresProductRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, errors.New("product ID is required")
	}

	query := `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, errors.New("failed to check if product exists")
	}

	return exists, nil
}

func (r *PostgresProductRepository) GetByCategory(ctx context.Context, categoryID string) ([]domain.Product, error) {
	if categoryID == "" {
		return nil, errors.New("category ID is required")
	}

	filter := domain.ProductFilter{
		CategoryID: categoryID,
		PageSize:   100, // Get up to 100 products
	}

	products, _, err := r.List(ctx, filter)
	return products, err
}

func (r *PostgresProductRepository) UpdateStock(ctx context.Context, productID string, newStock int) error {
	if productID == "" {
		return errors.New("product ID is required")
	}

	if newStock < 0 {
		return errors.New("stock cannot be negative")
	}

	query := `UPDATE products SET stock = $1, updated_at = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, newStock, time.Now(), productID)
	if err != nil {
		return errors.New("failed to update stock")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("failed to check update result")
	}

	if rowsAffected == 0 {
		return errors.New("product not found")
	}

	return nil
}

// Helper methods

func (r *PostgresProductRepository) buildWhereClause(filter domain.ProductFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.CategoryID != "" {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, filter.CategoryID)
		argIndex++
	}

	if filter.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price >= $%d", argIndex))
		args = append(args, *filter.MinPrice)
		argIndex++
	}

	if filter.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price <= $%d", argIndex))
		args = append(args, *filter.MaxPrice)
		argIndex++
	}

	if filter.InStock != nil && *filter.InStock {
		conditions = append(conditions, "stock > 0")
	}

	if filter.BikeType != "" {
		conditions = append(conditions, fmt.Sprintf("bike_type = $%d", argIndex))
		args = append(args, filter.BikeType)
		argIndex++
	}

	if filter.FrameSize != "" {
		conditions = append(conditions, fmt.Sprintf("frame_size = $%d", argIndex))
		args = append(args, filter.FrameSize)
		argIndex++
	}

	if filter.WheelSize != "" {
		conditions = append(conditions, fmt.Sprintf("wheel_size = $%d", argIndex))
		args = append(args, filter.WheelSize)
		argIndex++
	}

	if filter.Color != "" {
		conditions = append(conditions, fmt.Sprintf("color = $%d", argIndex))
		args = append(args, filter.Color)
		argIndex++
	}

	if filter.MaxWeight != nil {
		conditions = append(conditions, fmt.Sprintf("weight <= $%d", argIndex))
		args = append(args, *filter.MaxWeight)
		argIndex++
	}

	return strings.Join(conditions, " AND "), args
}

func (r *PostgresProductRepository) validateProduct(product domain.Product) error {
	if product.Name == "" {
		return errors.New("product name is required")
	}

	if product.Price < 0 {
		return errors.New("product price cannot be negative")
	}

	if product.Stock < 0 {
		return errors.New("product stock cannot be negative")
	}

	if product.CategoryID == "" {
		return errors.New("category ID is required")
	}

	return nil
}

// Helper functions for handling NULL values
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullFloat64(f float64) interface{} {
	if f == 0 {
		return nil
	}
	return f
}

func isForeignKeyError(err error) bool {
	// PostgreSQL specific - adjust for your database
	return err != nil && strings.Contains(err.Error(), "violates foreign key constraint")
}
