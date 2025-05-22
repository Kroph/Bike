package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type AuthService interface {
	GenerateToken(userID string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	GenerateVerificationToken(userID string) (string, error)
}

type Claims struct {
	UserID    string   `json:"user_id"`
	Role      UserRole `json:"role"`
	TokenType string   `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

// IsAdmin checks if user has admin role
func (c *Claims) IsAdmin() bool {
	return c.Role == UserRoleAdmin
}

// IsUser checks if user has user role
func (c *Claims) IsUser() bool {
	return c.Role == UserRoleUser
}

type authService struct {
	secretKey     string
	expiryMinutes int
}

func NewAuthService(secretKey string, expiryMinutes int) AuthService {
	return &authService{
		secretKey:     secretKey,
		expiryMinutes: expiryMinutes,
	}
}

func (s *authService) GenerateToken(userID string) (string, error) {
	expirationTime := time.Now().Add(time.Duration(s.expiryMinutes) * time.Minute)

	claims := &Claims{
		UserID:    userID,
		Role:      UserRoleUser,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

func (s *authService) GenerateTokenWithRole(userID string, role UserRole) (string, error) {
	expirationTime := time.Now().Add(time.Duration(s.expiryMinutes) * time.Minute)

	claims := &Claims{
		UserID:    userID,
		Role:      role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *authService) GenerateVerificationToken(userID string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID:    userID,
		Role:      UserRoleUser,
		TokenType: "verification",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}
