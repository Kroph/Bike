package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Services struct {
		Inventory struct {
			GrpcURL string
		}
	}
	NATS struct {
		URL string
	}
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{}

	config.Services.Inventory.GrpcURL = getEnv("INVENTORY_GRPC_URL", "localhost:50051")
	config.NATS.URL = getEnv("NATS_URL", "nats://localhost:4222")

	return config
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
