package config

import (
	"fmt"
	"os"
	"strconv"
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

	EmailEnabled           bool
	SMTPHost               string
	SMTPPort               int
	SMTPUsername           string
	SMTPPassword           string
	SMTPFrom               string
	SMTPFromName           string
	AdminNotificationEmail string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	appEnv := normalizeAppEnv(envOrDefault("APP_ENV", appEnvDevelopment))
	smtpPort, err := envIntOrDefault("SMTP_PORT", 587)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppName:                envOrDefault("APP_NAME", "Car Rental Web"),
		AppEnv:                 appEnv,
		IsProduction:           appEnv == appEnvProduction,
		AppPort:                envOrDefault("APP_PORT", "8080"),
		DatabaseURL:            envOrDefault("DATABASE_URL", "postgres://car_rental_user:car_rental_password@localhost:5432/car_rental_web?sslmode=disable"),
		SessionSecret:          envOrDefault("SESSION_SECRET", "change-me-in-development"),
		EmailEnabled:           envBoolOrDefault("EMAIL_ENABLED", false),
		SMTPHost:               strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:               smtpPort,
		SMTPUsername:           strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		SMTPPassword:           os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:               strings.TrimSpace(os.Getenv("SMTP_FROM")),
		SMTPFromName:           envOrDefault("SMTP_FROM_NAME", "Car Rental Web"),
		AdminNotificationEmail: strings.TrimSpace(os.Getenv("ADMIN_NOTIFICATION_EMAIL")),
	}

	if err := validateEmailConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func envBoolOrDefault(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	return value == "true" || value == "1" || value == "yes"
}

func envIntOrDefault(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer", key)
	}

	return parsed, nil
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

func validateEmailConfig(cfg Config) error {
	if !cfg.EmailEnabled {
		return nil
	}

	missing := make([]string, 0)
	if cfg.SMTPHost == "" {
		missing = append(missing, "SMTP_HOST")
	}
	if cfg.SMTPPort <= 0 {
		missing = append(missing, "SMTP_PORT")
	}
	if cfg.SMTPFrom == "" {
		missing = append(missing, "SMTP_FROM")
	}
	if cfg.AdminNotificationEmail == "" {
		missing = append(missing, "ADMIN_NOTIFICATION_EMAIL")
	}

	if len(missing) > 0 {
		return fmt.Errorf("email is enabled but required configuration is missing: %s", strings.Join(missing, ", "))
	}

	return nil
}
