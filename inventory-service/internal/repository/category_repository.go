package repository

import (
	"context"
	"database/sql"
	"errors"
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
}

type PostgresCategoryRepository struct {
	db *sql.DB
}

func NewPostgresCategoryRepository(db *sql.DB) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{
		db: db,
	}
}

func (r *PostgresCategoryRepository) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
	query := `
		INSERT INTO categories (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, created_at, updated_at
	`

	category.ID = uuid.New().String()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		category.ID,
		category.Name,
		category.Description,
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
		return domain.Category{}, err
	}

	return category, nil
}

func (r *PostgresCategoryRepository) GetByID(ctx context.Context, id string) (domain.Category, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

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
		return domain.Category{}, err
	}

	return category, nil
}

func (r *PostgresCategoryRepository) Update(ctx context.Context, category domain.Category) error {
	query := `
		UPDATE categories
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4
	`

	category.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		category.Name,
		category.Description,
		category.UpdatedAt,
		category.ID,
	)

	return err
}

func (r *PostgresCategoryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM categories WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresCategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM categories
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}
