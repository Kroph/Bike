package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DB struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
	}
	Server struct {
		Port     string
		GrpcPort string
	}
	Services struct {
		Inventory struct {
			GrpcURL string
		}
	}
	NATS struct {
		URL string
	}
	Redis struct {
		Addr     string
		Password string
		DB       int
	}
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{}

	config.DB.Host = getEnv("ORDERS_DB_HOST", "localhost")
	config.DB.Port = getEnv("ORDERS_DB_PORT", "5432")
	config.DB.User = getEnv("ORDERS_DB_USER", "postgres")
	config.DB.Password = getEnv("ORDERS_DB_PASSWORD", "postgres")
	config.DB.Name = getEnv("ORDERS_DB_NAME", "orders")

	config.Server.Port = getEnv("ORDERS_HTTP_PORT", "8081")
	config.Server.GrpcPort = getEnv("ORDERS_GRPC_PORT", "50052")

	config.Services.Inventory.GrpcURL = getEnv("INVENTORY_GRPC_URL", "localhost:50051")

	config.NATS.URL = getEnv("NATS_URL", "nats://localhost:4222")

	// Redis configuration
	config.Redis.Addr = getEnv("REDIS_ADDR", "localhost:6379")
	config.Redis.Password = getEnv("REDIS_PASSWORD", "")

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "3"))
	if err != nil {
		redisDB = 3
	}
	config.Redis.DB = redisDB

	return config
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
