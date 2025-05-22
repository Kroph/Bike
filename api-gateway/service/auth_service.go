package service

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type AuthService interface {
	ValidateToken(tokenString string) (*Claims, error)
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
	secretKey string
}

func NewAuthService(secretKey string, expiryMinutes int) AuthService {
	return &authService{
		secretKey: secretKey,
	}
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
