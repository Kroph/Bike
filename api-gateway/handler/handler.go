package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"api-gateway/middleware"
	"api-gateway/service"
)

type Handler struct {
	grpcClients  *service.GrpcClients
	authService  service.AuthService
	emailService service.EmailService
	redisClient  *redis.Client
}

func NewHandler(grpcClients *service.GrpcClients, authService service.AuthService,
	emailService service.EmailService, redisClient *redis.Client) *Handler {
	return &Handler{
		grpcClients:  grpcClients,
		authService:  authService,
		emailService: emailService,
		redisClient:  redisClient,
	}
}

func RegisterRoutes(router *gin.Engine, h *Handler) {
	// Public routes
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/register", h.RegisterUser)
		auth.POST("/login", h.Login)
	}

	// Protected routes - all require authentication
	api := router.Group("/api/v1")
	api.Use(service.AuthMiddleware(h.authService))
	{
		api.GET("/users/profile", h.GetUserProfile)
		api.POST("/users/verify-email", h.VerifyEmailCode)
		api.POST("/users/resend-verification", h.ResendVerificationCode)

		products := api.Group("/products")
		{
			products.GET("", h.ListProducts)
			products.GET("/:id", h.GetProduct)
			// Remove authService parameter - it's not needed
			products.POST("", middleware.RequireAdmin(), h.CreateProduct)
			products.PUT("/:id", middleware.RequireAdmin(), h.UpdateProduct)
			products.DELETE("/:id", middleware.RequireAdmin(), h.DeleteProduct)
		}

		categories := api.Group("/categories")
		{
			categories.GET("", h.ListCategories)
			categories.GET("/:id", h.GetCategory)
			// Remove authService parameter - it's not needed
			categories.POST("", middleware.RequireAdmin(), h.CreateCategory)
			categories.PUT("/:id", middleware.RequireAdmin(), h.UpdateCategory)
			categories.DELETE("/:id", middleware.RequireAdmin(), h.DeleteCategory)
		}

		orders := api.Group("/orders")
		{
			orders.POST("", h.CreateOrder)
			orders.GET("", h.ListUserOrders)
			orders.GET("/:id", h.GetOrder)
			orders.PATCH("/:id/status", h.UpdateOrderStatus)
		}
	}

	// Admin routes
	admin := router.Group("/api/v1/admin")
	admin.Use(service.AuthMiddleware(h.authService))
	admin.Use(middleware.RequireAdmin())
	{
		admin.GET("/orders", h.ListAllOrders)
		admin.GET("/orders/:id", h.GetAnyOrder)
		admin.PATCH("/orders/:id/status", h.AdminUpdateOrderStatus)
	}
}
