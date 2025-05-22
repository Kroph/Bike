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

	config.DB.Host = getEnv("DB_HOST", "localhost")
	config.DB.Port = getEnv("DB_PORT", "5432")
	config.DB.User = getEnv("DB_USER", "postgres")
	config.DB.Password = getEnv("DB_PASSWORD", "postgres")
	config.DB.Name = getEnv("DB_NAME", "inventory")

	config.Server.Port = getEnv("INVENTORY_HTTP_PORT", "8080")
	config.Server.GrpcPort = getEnv("INVENTORY_GRPC_PORT", "50051")

	// Redis configuration
	config.Redis.Addr = getEnv("REDIS_ADDR", "localhost:6379")
	config.Redis.Password = getEnv("REDIS_PASSWORD", "")

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "2"))
	if err != nil {
		redisDB = 2
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
