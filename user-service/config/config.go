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
	Auth struct {
		Secret        string
		ExpiryMinutes int
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

	config.DB.Host = getEnv("USER_DB_HOST", "localhost")
	config.DB.Port = getEnv("USER_DB_PORT", "5432")
	config.DB.User = getEnv("USER_DB_USER", "postgres")
	config.DB.Password = getEnv("USER_DB_PASSWORD", "postgres")
	config.DB.Name = getEnv("USER_DB_NAME", "users")

	config.Server.Port = getEnv("USER_HTTP_PORT", "8082")
	config.Server.GrpcPort = getEnv("USER_GRPC_PORT", "50053")

	config.Auth.Secret = getEnv("AUTH_SECRET", "your-secret-key-change-this-in-production")
	expiryStr := getEnv("AUTH_EXPIRY_MINUTES", "60")
	expiryMinutes, err := strconv.Atoi(expiryStr)
	if err != nil {
		expiryMinutes = 60
	}
	config.Auth.ExpiryMinutes = expiryMinutes

	// Redis configuration
	config.Redis.Addr = getEnv("REDIS_ADDR", "localhost:6379")
	config.Redis.Password = getEnv("REDIS_PASSWORD", "")

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "1"))
	if err != nil {
		redisDB = 1
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
