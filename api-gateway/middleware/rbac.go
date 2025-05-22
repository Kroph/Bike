package middleware

import (
	"api-gateway/service"

	"github.com/gin-gonic/gin"
)

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(401, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		role, ok := userRole.(service.UserRole)
		if !ok || role != service.UserRoleAdmin {
			c.JSON(403, gin.H{"error": "Admin privileges required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireAuthenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
