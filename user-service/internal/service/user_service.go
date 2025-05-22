package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"user-service/internal/cache"
	"user-service/internal/domain"
	"user-service/internal/repository"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	RegisterUser(ctx context.Context, user domain.User) (domain.User, error)
	AuthenticateUser(ctx context.Context, email, password string) (string, domain.User, error)
	GetUserProfile(ctx context.Context, id string) (domain.User, error)
	GenerateVerificationCode(ctx context.Context, userID string) (string, error)
	VerifyEmailCode(ctx context.Context, userID, code string) error
}

type userService struct {
	userRepo    repository.UserRepository
	jwtSecret   string
	jwtDuration time.Duration
	cache       cache.Cache
}

type Claims struct {
	UserID    string          `json:"user_id"`
	Role      domain.UserRole `json:"role"`
	TokenType string          `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

func NewUserService(userRepo repository.UserRepository, jwtSecret string, jwtExpiryMinutes int, cache cache.Cache) UserService {
	return &userService{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		jwtDuration: time.Duration(jwtExpiryMinutes) * time.Minute,
		cache:       cache,
	}
}

func (s *userService) RegisterUser(ctx context.Context, user domain.User) (domain.User, error) {
	_, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err == nil {
		return domain.User{}, errors.New("user with this email already exists")
	}

	if user.Role == "" || user.Role == domain.UserRoleAdmin {
		user.Role = domain.UserRoleUser
	}

	return s.userRepo.Create(ctx, user)
}

func (s *userService) AuthenticateUser(ctx context.Context, email, password string) (string, domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", domain.User{}, errors.New("invalid email or password")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", domain.User{}, errors.New("invalid email or password")
	}

	// Generate JWT token with role
	expirationTime := time.Now().Add(s.jwtDuration)
	claims := &Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", domain.User{}, err
	}

	// Clear password before returning
	user.Password = ""

	return tokenString, user, nil
}

func (s *userService) GetUserProfile(ctx context.Context, id string) (domain.User, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:profile:%s", id)
	var cachedUser domain.User

	err := s.cache.Get(ctx, cacheKey, &cachedUser)
	if err == nil {
		log.Printf("Cache hit for user profile ID: %s", id)
		return cachedUser, nil
	}

	if err != cache.ErrCacheMiss {
		log.Printf("Cache error for user profile ID %s: %v", id, err)
	}

	// If not in cache, get from database
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	// Clear password before caching
	user.Password = ""

	// Store in cache with 15-minute TTL
	if err := s.cache.Set(ctx, cacheKey, user, 15*time.Minute); err != nil {
		log.Printf("Failed to cache user profile ID %s: %v", id, err)
	}

	return user, nil
}

func (s *userService) GenerateVerificationCode(ctx context.Context, userID string) (string, error) {
	// Generate 6-digit code
	code, err := s.generateSixDigitCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate verification code: %v", err)
	}

	// Store code in cache with 10-minute expiry
	cacheKey := fmt.Sprintf("verification_code:%s", userID)
	err = s.cache.Set(ctx, cacheKey, code, 10*time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed to store verification code: %v", err)
	}

	log.Printf("Generated verification code for user %s: %s", userID, code)
	return code, nil
}

func (s *userService) VerifyEmailCode(ctx context.Context, userID, code string) error {
	// Get stored code from cache
	cacheKey := fmt.Sprintf("verification_code:%s", userID)
	var storedCode string

	err := s.cache.Get(ctx, cacheKey, &storedCode)
	if err != nil {
		if err == cache.ErrCacheMiss {
			return errors.New("verification code expired or not found")
		}
		return fmt.Errorf("failed to retrieve verification code: %v", err)
	}

	// Compare codes
	if storedCode != code {
		return errors.New("invalid verification code")
	}

	// Mark email as verified
	verificationKey := fmt.Sprintf("email_verified:%s", userID)
	err = s.cache.Set(ctx, verificationKey, "true", 0) // No expiration
	if err != nil {
		return fmt.Errorf("failed to mark email as verified: %v", err)
	}

	// Delete the verification code
	s.cache.Delete(ctx, cacheKey)

	log.Printf("Email verified successfully for user: %s", userID)
	return nil
}

func (s *userService) generateSixDigitCode() (string, error) {
	min := big.NewInt(100000)
	max := big.NewInt(999999)

	n, err := rand.Int(rand.Reader, new(big.Int).Sub(max, min))
	if err != nil {
		return "", err
	}

	code := new(big.Int).Add(min, n)
	return code.String(), nil
}
