package repository

import (
	"context"
	"database/sql"
	"errors"
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
