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
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) UserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	// Validate input
	if user.Email == "" || user.Username == "" {
		return domain.User{}, errors.New("email and username are required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, errors.New("failed to hash password")
	}

	// Default role to user if not specified
	if user.Role == "" {
		user.Role = domain.UserRoleUser
	}

	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, username, email, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, username, email, role, created_at, updated_at`

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

	if err != nil {
		// Handle specific database errors
		if isDuplicateKeyError(err) {
			return domain.User{}, errors.New("user with this email or username already exists")
		}
		return domain.User{}, errors.New("failed to create user")
	}

	user.Role = domain.UserRole(roleStr)
	user.Password = "" // Clear password for security

	return user, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	if id == "" {
		return domain.User{}, errors.New("user ID is required")
	}

	query := `
		SELECT id, username, email, password, role, created_at, updated_at
		FROM users
		WHERE id = $1`

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
		return domain.User{}, errors.New("failed to get user")
	}

	user.Role = domain.UserRole(roleStr)
	return user, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	if email == "" {
		return domain.User{}, errors.New("email is required")
	}

	query := `
		SELECT id, username, email, password, role, created_at, updated_at
		FROM users
		WHERE email = $1`

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
		return domain.User{}, errors.New("failed to get user")
	}

	user.Role = domain.UserRole(roleStr)
	return user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user domain.User) error {
	if user.ID == "" {
		return errors.New("user ID is required")
	}

	query := `
		UPDATE users
		SET username = $1, email = $2, role = $3, updated_at = $4
		WHERE id = $5`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.Email,
		string(user.Role),
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		if isDuplicateKeyError(err) {
			return errors.New("username or email already exists")
		}
		return errors.New("failed to update user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("failed to check update result")
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *PostgresUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, errors.New("email is required")
	}

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, errors.New("failed to check if user exists")
	}

	return exists, nil
}

// Helper function to check for duplicate key errors
func isDuplicateKeyError(err error) bool {
	// This is PostgreSQL specific - you might want to use a more robust solution
	return err != nil && (
	// Check for various duplicate key error patterns
	// This is simplified - in production, use a proper error checking library
	err.Error() == "duplicate key value violates unique constraint" ||
		err.Error() == "UNIQUE constraint failed")
}
