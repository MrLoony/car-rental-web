package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName       string
	AppEnv        string
	IsProduction  bool
	AppPort       string
	DatabaseURL   string
	SessionSecret string
}

func Load() Config {
	_ = godotenv.Load()

	appEnv := normalizeAppEnv(envOrDefault("APP_ENV", appEnvDevelopment))

	return Config{
		AppName:       envOrDefault("APP_NAME", "Car Rental Web"),
		AppEnv:        appEnv,
		IsProduction:  appEnv == appEnvProduction,
		AppPort:       envOrDefault("APP_PORT", "8080"),
		DatabaseURL:   envOrDefault("DATABASE_URL", "postgres://car_rental_user:car_rental_password@localhost:5432/car_rental_web?sslmode=disable"),
		SessionSecret: envOrDefault("SESSION_SECRET", "change-me-in-development"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

const (
	appEnvDevelopment = "development"
	appEnvProduction  = "production"
)

func normalizeAppEnv(value string) string {
	if strings.TrimSpace(value) == appEnvProduction {
		return appEnvProduction
	}

	return appEnvDevelopment
}
