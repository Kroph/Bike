package service_test

import (
	"testing"
	"time"

	"api-gateway/service"

	"github.com/golang-jwt/jwt/v4"
)

func generateTestToken(secret string, userID string, role service.UserRole) string {
	claims := &service.Claims{
		UserID:    userID,
		Role:      role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

func TestValidateToken_Success(t *testing.T) {
	secret := "test-secret"
	authService := service.NewAuthService(secret, 60)
	token := generateTestToken(secret, "user123", service.UserRoleUser)

	claims, err := authService.ValidateToken(token)
	if err != nil {
		t.Errorf("Expected valid token, got error: %v", err)
	}
	if claims.UserID != "user123" {
		t.Errorf("Expected userID 'user123', got %s", claims.UserID)
	}
	if claims.Role != service.UserRoleUser {
		t.Errorf("Expected role 'user', got %s", claims.Role)
	}
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	goodSecret := "correct-secret"
	badSecret := "wrong-secret"
	token := generateTestToken(goodSecret, "user123", service.UserRoleUser)
	authService := service.NewAuthService(badSecret, 60)

	_, err := authService.ValidateToken(token)
	if err == nil {
		t.Errorf("Expected error due to invalid signature, got none")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	secret := "test-secret"
	claims := &service.Claims{
		UserID:    "user123",
		Role:      service.UserRoleUser,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	authService := service.NewAuthService(secret, 60)

	_, err := authService.ValidateToken(tokenStr)
	if err == nil {
		t.Errorf("Expected error due to expired token, got none")
	}
}

func TestClaims_IsAdmin(t *testing.T) {
	claims := &service.Claims{Role: service.UserRoleAdmin}
	if !claims.IsAdmin() {
		t.Errorf("Expected IsAdmin to return true for admin role")
	}
	if claims.IsUser() {
		t.Errorf("Expected IsUser to return false for admin role")
	}
}

func TestClaims_IsUser(t *testing.T) {
	claims := &service.Claims{Role: service.UserRoleUser}
	if !claims.IsUser() {
		t.Errorf("Expected IsUser to return true for user role")
	}
	if claims.IsAdmin() {
		t.Errorf("Expected IsAdmin to return false for user role")
	}
}
