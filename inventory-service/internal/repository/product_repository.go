package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
}

type PostgresProductRepository struct {
	db *sql.DB
}

func NewPostgresProductRepository(db *sql.DB) *PostgresProductRepository {
	return &PostgresProductRepository{
		db: db,
	}
}

func (r *PostgresProductRepository) Create(ctx context.Context, product domain.Product) (domain.Product, error) {
	query := `
		INSERT INTO products (id, name, description, price, stock, category_id, 
		                      frame_size, wheel_size, color, weight, bike_type, 
		                      created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, name, description, price, stock, category_id, 
		          frame_size, wheel_size, color, weight, bike_type, 
		          created_at, updated_at
	`

	product.ID = uuid.New().String()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.CategoryID,
		product.FrameSize,
		product.WheelSize,
		product.Color,
		product.Weight,
		product.BikeType,
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
		return domain.Product{}, err
	}

	return product, nil
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id string) (domain.Product, error) {
	query := `
        SELECT id, name, description, price, stock, category_id, 
               frame_size, wheel_size, color, weight, bike_type, 
               created_at, updated_at
        FROM products
        WHERE id = $1
    `

	var product domain.Product
	var frameSize, wheelSize, color, bikeType sql.NullString
	var weight sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.CategoryID,
		&frameSize,
		&wheelSize,
		&color,
		&weight,
		&bikeType,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Product{}, errors.New("product not found")
		}
		return domain.Product{}, err
	}

	// Handle NULL values for strings
	product.FrameSize = getStringValue(frameSize)
	product.WheelSize = getStringValue(wheelSize)
	product.Color = getStringValue(color)
	product.BikeType = getStringValue(bikeType)

	// Handle NULL values for floats
	if weight.Valid {
		product.Weight = weight.Float64
	} else {
		product.Weight = 0.0
	}

	return product, nil
}

func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func (r *PostgresProductRepository) Update(ctx context.Context, product domain.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, stock = $4, category_id = $5,
		    frame_size = $6, wheel_size = $7, color = $8, weight = $9, bike_type = $10, 
		    updated_at = $11
		WHERE id = $12
	`

	product.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.CategoryID,
		product.FrameSize,
		product.WheelSize,
		product.Color,
		product.Weight,
		product.BikeType,
		product.UpdatedAt,
		product.ID,
	)

	return err
}

func (r *PostgresProductRepository) List(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, int, error) {
	baseQuery := `
        SELECT id, name, description, price, stock, category_id, 
               frame_size, wheel_size, color, weight, bike_type, 
               created_at, updated_at
        FROM products
        WHERE 1=1
    `

	countQuery := `
        SELECT COUNT(*)
        FROM products
        WHERE 1=1
    `

	var conditions string
	var args []interface{}
	var argIndex int = 1

	if filter.CategoryID != "" {
		conditions += fmt.Sprintf(" AND category_id = $%d", argIndex)
		args = append(args, filter.CategoryID)
		argIndex++
	}

	if filter.MinPrice != nil {
		conditions += fmt.Sprintf(" AND price >= $%d", argIndex)
		args = append(args, *filter.MinPrice)
		argIndex++
	}

	if filter.MaxPrice != nil {
		conditions += fmt.Sprintf(" AND price <= $%d", argIndex)
		args = append(args, *filter.MaxPrice)
		argIndex++
	}

	if filter.InStock != nil && *filter.InStock {
		conditions += " AND stock > 0"
	}

	// Bicycle-specific filters
	if filter.BikeType != "" {
		conditions += fmt.Sprintf(" AND bike_type = $%d", argIndex)
		args = append(args, filter.BikeType)
		argIndex++
	}

	if filter.FrameSize != "" {
		conditions += fmt.Sprintf(" AND frame_size = $%d", argIndex)
		args = append(args, filter.FrameSize)
		argIndex++
	}

	if filter.WheelSize != "" {
		conditions += fmt.Sprintf(" AND wheel_size = $%d", argIndex)
		args = append(args, filter.WheelSize)
		argIndex++
	}

	if filter.Color != "" {
		conditions += fmt.Sprintf(" AND color = $%d", argIndex)
		args = append(args, filter.Color)
		argIndex++
	}

	if filter.MaxWeight != nil {
		conditions += fmt.Sprintf(" AND weight <= $%d", argIndex)
		args = append(args, *filter.MaxWeight)
		argIndex++
	}

	// Pagination
	limit := 10
	offset := 0

	if filter.PageSize > 0 {
		limit = filter.PageSize
	}

	if filter.Page > 0 {
		offset = (filter.Page - 1) * limit
	}

	query := baseQuery + conditions + fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute the query to get the products
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Process the result rows
	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		var frameSize, wheelSize, color, bikeType sql.NullString
		var weight sql.NullFloat64

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.CategoryID,
			&frameSize,
			&wheelSize,
			&color,
			&weight,
			&bikeType,
			&product.CreatedAt,
			&product.UpdatedAt,
		)

		if err != nil {
			return nil, 0, err
		}

		// Handle NULL values for strings
		product.FrameSize = getStringValue(frameSize)
		product.WheelSize = getStringValue(wheelSize)
		product.Color = getStringValue(color)
		product.BikeType = getStringValue(bikeType)

		// Handle NULL values for floats
		if weight.Valid {
			product.Weight = weight.Float64
		} else {
			product.Weight = 0.0
		}

		products = append(products, product)
	}

	// Get the total count for pagination
	var total int
	err = r.db.QueryRowContext(ctx, countQuery+conditions, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
