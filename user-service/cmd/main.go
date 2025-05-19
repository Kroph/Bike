package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	"proto/user"
	"user-service/config"
	"user-service/internal/cache"
	"user-service/internal/handler"
	"user-service/internal/repository"
	"user-service/internal/service"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.LoadConfig()

	// Connect to database
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Check DB connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisCache.Close()

	// Initialize repositories
	userRepo := repository.NewPostgresUserRepository(db)

	// Initialize services with cache
	userService := service.NewUserService(userRepo, cfg.Auth.Secret, cfg.Auth.ExpiryMinutes, redisCache)

	// Initialize gRPC server
	lis, err := net.Listen("tcp", ":"+cfg.Server.GrpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Register user service handler
	userHandler := handler.NewUserGrpcHandler(userService)
	user.RegisterUserServiceServer(grpcServer, userHandler)

	// Enable reflection for tools like grpcurl
	reflection.Register(grpcServer)

	log.Printf("User Service gRPC server starting on port %s", cfg.Server.GrpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
