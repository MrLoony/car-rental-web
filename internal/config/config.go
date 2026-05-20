package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName string
	AppEnv  string
	AppPort string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		AppName: envOrDefault("APP_NAME", "Car Rental Web"),
		AppEnv:  envOrDefault("APP_ENV", "development"),
		AppPort: envOrDefault("APP_PORT", "8080"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
