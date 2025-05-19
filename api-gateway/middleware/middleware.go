package middleware

import (
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"api-gateway/service"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("RequestID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		timestamp := time.Now()
		latency := timestamp.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		requestID, _ := c.Get("RequestID")

		if raw != "" {
			path = path + "?" + raw
		}

		log.Printf("[API-GATEWAY] %s | %3d | %13v | %15s | %s | %s | %s",
			timestamp.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
			requestID,
		)
	}
}

func Telemetry() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		responseTime := time.Since(startTime)

		log.Printf("[TELEMETRY] Path: %s, Method: %s, Status: %d, ResponseTime: %v",
			c.Request.URL.Path,
			c.Request.Method,
			c.Writer.Status(),
			responseTime,
		)
	}
}

func Authentication(authService service.AuthService) gin.HandlerFunc {
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

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
