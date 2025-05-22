package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"user-service/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter domain.UserFilter) ([]domain.User, int, error)
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}

	// Default role to user if not specified
	if user.Role == "" {
		user.Role = domain.UserRoleUser
	}

	query := `
		INSERT INTO users (id, username, email, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, username, email, role, created_at, updated_at
	`

	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	var roleStr string
	err = r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.Username,
		user.Email,
		string(hashedPassword),
		string(user.Role),
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&roleStr,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	user.Role = domain.UserRole(roleStr)
	user.Password = "" // Clear password

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	query := `
		SELECT id, username, email, password, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&roleStr,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, errors.New("user not found")
		}
		return domain.User{}, err
	}

	user.Role = domain.UserRole(roleStr)
	return user, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `
		SELECT id, username, email, password, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&roleStr,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, errors.New("user not found")
		}
		return domain.User{}, err
	}

	user.Role = domain.UserRole(roleStr)
	return user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user domain.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, role = $3, updated_at = $4
		WHERE id = $5
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.Email,
		string(user.Role),
		user.UpdatedAt,
		user.ID,
	)

	return err
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresUserRepository) List(ctx context.Context, filter domain.UserFilter) ([]domain.User, int, error) {
	baseQuery := `
		SELECT id, username, email, role, created_at, updated_at
		FROM users
		WHERE 1=1
	`

	countQuery := `
		SELECT COUNT(*)
		FROM users
		WHERE 1=1
	`

	var conditions string
	var args []interface{}
	var argIndex int = 1

	if filter.Email != "" {
		conditions += fmt.Sprintf(" AND email = $%d", argIndex)
		args = append(args, filter.Email)
		argIndex++
	}

	if filter.Username != "" {
		conditions += fmt.Sprintf(" AND username = $%d", argIndex)
		args = append(args, filter.Username)
		argIndex++
	}

	if filter.Role != "" {
		conditions += fmt.Sprintf(" AND role = $%d", argIndex)
		args = append(args, string(filter.Role))
		argIndex++
	}

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

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		var roleStr string
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&roleStr,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		user.Role = domain.UserRole(roleStr)
		users = append(users, user)
	}

	var total int
	err = r.db.QueryRowContext(ctx, countQuery+conditions, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
