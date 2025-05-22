package middleware

import (
	"strings"

	"api-gateway/service"

	"github.com/gin-gonic/gin"
)

func RequireRole(authService service.AuthService, requiredRole service.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "Authorization header must be in format: Bearer {token}"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims.Role != requiredRole {
			c.JSON(403, gin.H{"error": "Insufficient privileges"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// RequireAdmin middleware that checks if user is admin
func RequireAdmin(authService service.AuthService) gin.HandlerFunc {
	return RequireRole(authService, service.UserRoleAdmin)
}

// RequireUserOrAdmin middleware that allows both user and admin
func RequireUserOrAdmin(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "Authorization header must be in format: Bearer {token}"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Accept both user and admin roles
		if claims.Role != service.UserRoleUser && claims.Role != service.UserRoleAdmin {
			c.JSON(403, gin.H{"error": "Invalid user role"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}
