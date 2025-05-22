package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"api-gateway/middleware"
	"api-gateway/service"
	inventorypb "proto/inventory"
	orderpb "proto/order"
	userpb "proto/user"
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

// User handlers

func (h *Handler) RegisterUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.grpcClients.RegisterUser(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Authenticate user to get a proper token from User Service
	authResponse, err := h.grpcClients.AuthenticateUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration successful but failed to generate token"})
		return
	}

	// Generate verification code via user service
	code, err := h.grpcClients.GenerateVerificationCode(c.Request.Context(), user.Id)
	if err != nil {
		log.Printf("Failed to generate verification code: %v", err)
		// Continue without verification email
	} else {
		// Send verification email with code
		err = h.emailService.SendEmailVerificationCode(req.Email, req.Username, code)
		if err != nil {
			log.Printf("Failed to send verification email: %v", err)
		} else {
			log.Printf("Verification email sent successfully to %s with code: %s", req.Email, code)
		}
	}

	// Map proto role to string for response
	var roleStr string
	switch user.Role {
	case userpb.UserRole_ADMIN:
		roleStr = "admin"
	case userpb.UserRole_USER:
		fallthrough
	default:
		roleStr = "user"
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully. Please check your email for a 6-digit verification code.",
		"user": gin.H{
			"id":       user.Id,
			"username": user.Username,
			"email":    user.Email,
			"role":     roleStr,
		},
		"token": authResponse.Token,
	})
}

func (h *Handler) VerifyEmailCode(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required,len=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Verify code via user service
	success, err := h.grpcClients.VerifyEmailCode(c.Request.Context(), userID.(string), req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !success {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired verification code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}

func (h *Handler) ResendVerificationCode(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get user profile to get email
	userProfile, err := h.grpcClients.GetUserProfile(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	// Generate new verification code
	code, err := h.grpcClients.GenerateVerificationCode(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate verification code"})
		return
	}

	// Send verification email with new code
	err = h.emailService.SendEmailVerificationCode(userProfile.Email, userProfile.Username, code)
	if err != nil {
		log.Printf("Failed to send verification email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification code sent successfully",
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authResponse, err := h.grpcClients.AuthenticateUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if email is verified
	ctx := c.Request.Context()
	verificationKey := fmt.Sprintf("email_verified:%s", authResponse.UserId)

	verified, err := h.redisClient.Get(ctx, verificationKey).Result()
	if err != nil && err != redis.Nil {
		log.Printf("Failed to check email verification status: %v", err)
		// Continue with login even if we can't check verification status
	}

	// Map proto role to string for response
	var roleStr string
	switch authResponse.Role {
	case userpb.UserRole_ADMIN:
		roleStr = "admin"
	case userpb.UserRole_USER:
		fallthrough
	default:
		roleStr = "user"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":             authResponse.UserId,
			"username":       authResponse.Username,
			"email":          authResponse.Email,
			"role":           roleStr,
			"email_verified": verified == "true",
		},
		"token": authResponse.Token,
	})
}

func (h *Handler) GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	profile, err := h.grpcClients.GetUserProfile(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Map proto role to string for response
	var roleStr string
	switch profile.Role {
	case userpb.UserRole_ADMIN:
		roleStr = "admin"
	case userpb.UserRole_USER:
		fallthrough
	default:
		roleStr = "user"
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         profile.Id,
			"username":   profile.Username,
			"email":      profile.Email,
			"role":       roleStr,
			"created_at": profile.CreatedAt,
		},
	})
}

// Product handlers

func (h *Handler) CreateProduct(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		Price       float64 `json:"price" binding:"required,gt=0"`
		Stock       int32   `json:"stock" binding:"required,gte=0"`
		CategoryID  string  `json:"category_id" binding:"required"`
		FrameSize   string  `json:"frame_size"`
		WheelSize   string  `json:"wheel_size"`
		Color       string  `json:"color"`
		Weight      float64 `json:"weight"`
		BikeType    string  `json:"bike_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.grpcClients.CreateProduct(c.Request.Context(), &inventorypb.CreateProductRequest{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryId:  req.CategoryID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Stock       int32   `json:"stock"`
		CategoryID  string  `json:"category_id"`
		FrameSize   string  `json:"frame_size"`
		WheelSize   string  `json:"wheel_size"`
		Color       string  `json:"color"`
		Weight      float64 `json:"weight"`
		BikeType    string  `json:"bike_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.grpcClients.UpdateProduct(c.Request.Context(), &inventorypb.UpdateProductRequest{
		Id:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryId:  req.CategoryID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *Handler) ListProducts(c *gin.Context) {
	var filter inventorypb.ProductFilter

	if categoryID := c.Query("category_id"); categoryID != "" {
		filter.CategoryId = categoryID
	}

	if minPrice := c.Query("min_price"); minPrice != "" {
		if minPriceFloat, err := strconv.ParseFloat(minPrice, 64); err == nil {
			filter.MinPrice = minPriceFloat
		}
	}

	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if maxPriceFloat, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			filter.MaxPrice = maxPriceFloat
		}
	}

	if inStock := c.Query("in_stock"); inStock == "true" {
		filter.InStock = true
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	filter.Page = int32(page)
	filter.PageSize = int32(pageSize)

	response, err := h.grpcClients.ListProducts(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetProduct(c *gin.Context) {
	id := c.Param("id")

	product, err := h.grpcClients.GetProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *Handler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	response, err := h.grpcClients.DeleteProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
	})
}

// Category handlers

func (h *Handler) CreateCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.grpcClients.CreateCategory(c.Request.Context(), &inventorypb.CreateCategoryRequest{
		Name:        req.Name,
		Description: req.Description,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

func (h *Handler) GetCategory(c *gin.Context) {
	id := c.Param("id")

	category, err := h.grpcClients.GetCategory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.grpcClients.UpdateCategory(c.Request.Context(), &inventorypb.UpdateCategoryRequest{
		Id:          id,
		Name:        req.Name,
		Description: req.Description,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	response, err := h.grpcClients.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
	})
}

func (h *Handler) ListCategories(c *gin.Context) {
	response, err := h.grpcClients.ListCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Order handlers

func (h *Handler) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Items []struct {
			ProductID string  `json:"product_id" binding:"required"`
			Name      string  `json:"name" binding:"required"`
			Price     float64 `json:"price" binding:"required,gt=0"`
			Quantity  int32   `json:"quantity" binding:"required,gt=0"`
			FrameSize string  `json:"frame_size"`
			WheelSize string  `json:"wheel_size"`
			Color     string  `json:"color"`
			BikeType  string  `json:"bike_type"`
		} `json:"items" binding:"required,dive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orderItems []*orderpb.OrderItemRequest
	for _, item := range req.Items {
		orderItems = append(orderItems, &orderpb.OrderItemRequest{
			ProductId: item.ProductID,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  item.Quantity,
		})
	}

	// Check stock availability
	var productQuantities []*inventorypb.ProductQuantity
	for _, item := range req.Items {
		productQuantities = append(productQuantities, &inventorypb.ProductQuantity{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	stockCheck, err := h.grpcClients.CheckStock(c.Request.Context(), productQuantities)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check stock: " + err.Error()})
		return
	}

	if !stockCheck.Available {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "Some bicycles are out of stock",
			"unavailable_items": stockCheck.UnavailableItems,
		})
		return
	}

	order, err := h.grpcClients.CreateOrder(c.Request.Context(), &orderpb.CreateOrderRequest{
		UserId: userID.(string),
		Items:  orderItems,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get user details for the email
	userProfile, err := h.grpcClients.GetUserProfile(c.Request.Context(), userID.(string))
	if err != nil {
		log.Printf("Failed to get user profile for email notification: %v", err)
	} else {
		// Prepare order details for email
		orderDetails := map[string]interface{}{
			"Items": make([]map[string]interface{}, 0, len(order.Items)),
			"Total": order.Total,
		}

		for _, item := range order.Items {
			orderDetails["Items"] = append(orderDetails["Items"].([]map[string]interface{}), map[string]interface{}{
				"Name":      item.Name,
				"Price":     item.Price,
				"Quantity":  item.Quantity,
				"Subtotal":  item.Price * float64(item.Quantity),
				"FrameSize": req.Items[0].FrameSize,
				"WheelSize": req.Items[0].WheelSize,
				"Color":     req.Items[0].Color,
				"BikeType":  req.Items[0].BikeType,
			})
		}

		// Send order confirmation email
		err = h.emailService.SendOrderConfirmation(userProfile.Email, order.Id, orderDetails)
		if err != nil {
			log.Printf("Failed to send order confirmation email: %v", err)
			// Continue with order response even if email fails
		}
	}

	c.JSON(http.StatusCreated, order)
}

func (h *Handler) GetOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")

	order, err := h.grpcClients.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Ensure the order belongs to the authenticated user (unless admin)
	userRole, _ := c.Get("user_role")
	if userRole != service.UserRoleAdmin && order.UserId != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")

	// First, check if the order belongs to the user (unless admin)
	order, err := h.grpcClients.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	userRole, _ := c.Get("user_role")
	if userRole != service.UserRoleAdmin && order.UserId != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var status orderpb.OrderStatus
	switch req.Status {
	case "pending":
		status = orderpb.OrderStatus_PENDING
	case "paid":
		status = orderpb.OrderStatus_PAID
	case "shipped":
		status = orderpb.OrderStatus_SHIPPED
	case "delivered":
		status = orderpb.OrderStatus_DELIVERED
	case "cancelled":
		status = orderpb.OrderStatus_CANCELLED
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	updatedOrder, err := h.grpcClients.UpdateOrderStatus(c.Request.Context(), &orderpb.UpdateOrderStatusRequest{
		Id:     id,
		Status: status,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedOrder)
}

func (h *Handler) ListUserOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	response, err := h.grpcClients.GetUserOrders(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Admin-only handlers

// ListAllOrders - Admin only: List
func (h *Handler) ListAllOrders(c *gin.Context) {
	// Parse query parameters for filtering
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	userID := c.Query("user_id")

	var filter orderpb.OrderFilter
	filter.Page = int32(page)
	filter.PageSize = int32(pageSize)

	if userID != "" {
		filter.UserId = userID
	}

	if status != "" {
		switch status {
		case "pending":
			filter.Status = orderpb.OrderStatus_PENDING
		case "paid":
			filter.Status = orderpb.OrderStatus_PAID
		case "shipped":
			filter.Status = orderpb.OrderStatus_SHIPPED
		case "delivered":
			filter.Status = orderpb.OrderStatus_DELIVERED
		case "cancelled":
			filter.Status = orderpb.OrderStatus_CANCELLED
		}
	}

	response, err := h.grpcClients.ListOrders(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetAnyOrder - Admin only: Get any order by ID
func (h *Handler) GetAnyOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.grpcClients.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// AdminUpdateOrderStatus - Admin only: Update any order's status
func (h *Handler) AdminUpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var status orderpb.OrderStatus
	switch req.Status {
	case "pending":
		status = orderpb.OrderStatus_PENDING
	case "paid":
		status = orderpb.OrderStatus_PAID
	case "shipped":
		status = orderpb.OrderStatus_SHIPPED
	case "delivered":
		status = orderpb.OrderStatus_DELIVERED
	case "cancelled":
		status = orderpb.OrderStatus_CANCELLED
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	updatedOrder, err := h.grpcClients.UpdateOrderStatus(c.Request.Context(), &orderpb.UpdateOrderStatusRequest{
		Id:     id,
		Status: status,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedOrder)
}

// Helper functions

func (h *Handler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification token is required"})
		return
	}

	// Validate the token
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired verification token"})
		return
	}

	// Check if it's a verification token
	if claims.TokenType != "verification" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token type"})
		return
	}

	// Store verification status in Redis
	ctx := c.Request.Context()
	verificationKey := fmt.Sprintf("email_verified:%s", claims.UserID)

	// Store with no expiration
	err = h.redisClient.Set(ctx, verificationKey, "true", 0).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email has been successfully verified",
	})
}

func (h *Handler) TestEmailCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[TEST-EMAIL-CODE] Starting email code test to: %s", req.Email)

	// Generate a test 6-digit code
	testCode := "123456"

	// Test email sending with code
	err := h.emailService.SendEmailVerificationCode(req.Email, "Test User", testCode)
	if err != nil {
		log.Printf("[TEST-EMAIL-CODE] Failed to send test email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send email",
			"details": err.Error(),
		})
		return
	}

	log.Printf("[TEST-EMAIL-CODE] Test email with code sent successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Test email with verification code sent successfully",
		"email":   req.Email,
		"code":    testCode,
	})
}

// Helper function to check if user is admin
func (h *Handler) isUserAdmin(c *gin.Context) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}

	role, ok := userRole.(service.UserRole)
	if !ok {
		return false
	}

	return role == service.UserRoleAdmin
}

func RegisterRoutes(router *gin.Engine, h *Handler) {
	// Public routes
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/register", h.RegisterUser)
		auth.POST("/login", h.Login)
	}

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

			products.POST("", middleware.RequireAdmin(h.authService), h.CreateProduct)
			products.PUT("/:id", middleware.RequireAdmin(h.authService), h.UpdateProduct)
			products.DELETE("/:id", middleware.RequireAdmin(h.authService), h.DeleteProduct)
		}

		categories := api.Group("/categories")
		{
			categories.GET("", h.ListCategories)
			categories.GET("/:id", h.GetCategory)

			categories.POST("", middleware.RequireAdmin(h.authService), h.CreateCategory)
			categories.PUT("/:id", middleware.RequireAdmin(h.authService), h.UpdateCategory)
			categories.DELETE("/:id", middleware.RequireAdmin(h.authService), h.DeleteCategory)
		}

		orders := api.Group("/orders")
		{
			orders.POST("", h.CreateOrder)
			orders.GET("", h.ListUserOrders)
			orders.GET("/:id", h.GetOrder)
			orders.PATCH("/:id/status", h.UpdateOrderStatus)
		}
	}

	admin := router.Group("/api/v1/admin")
	admin.Use(middleware.RequireAdmin(h.authService))
	{
		admin.GET("/orders", h.ListAllOrders)
		admin.GET("/orders/:id", h.GetAnyOrder)
		admin.PATCH("/orders/:id/status", h.AdminUpdateOrderStatus)
	}

	// Test routes
	router.POST("/api/v1/test-email-code", h.TestEmailCode)
}
