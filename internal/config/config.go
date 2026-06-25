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
	BaseURL       string
	DatabaseURL   string
	SessionSecret string
	AdminEmail    string
	AdminPassword string

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
		AppPort:                envOrDefault("PORT", envOrDefault("APP_PORT", "8080")),
		BaseURL:                strings.TrimRight(strings.TrimSpace(os.Getenv("BASE_URL")), "/"),
		DatabaseURL:            envOrDefault("DATABASE_URL", defaultDatabaseURL),
		SessionSecret:          envOrDefault("SESSION_SECRET", defaultSessionSecret),
		AdminEmail:             strings.TrimSpace(os.Getenv("ADMIN_EMAIL")),
		AdminPassword:          os.Getenv("ADMIN_PASSWORD"),
		EmailEnabled:           envBoolOrDefault("EMAIL_ENABLED", false),
		SMTPHost:               strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:               smtpPort,
		SMTPUsername:           strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		SMTPPassword:           os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:               strings.TrimSpace(os.Getenv("SMTP_FROM")),
		SMTPFromName:           envOrDefault("SMTP_FROM_NAME", "Car Rental Web"),
		AdminNotificationEmail: strings.TrimSpace(os.Getenv("ADMIN_NOTIFICATION_EMAIL")),
	}

	if err := validateProductionConfig(cfg); err != nil {
		return Config{}, err
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

	defaultDatabaseURL   = "postgres://car_rental_user:car_rental_password@localhost:5432/car_rental_web?sslmode=disable"
	defaultSessionSecret = "change-me-in-development"
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

func validateProductionConfig(cfg Config) error {
	if !cfg.IsProduction {
		return nil
	}

	missing := make([]string, 0)
	if cfg.DatabaseURL == "" || cfg.DatabaseURL == defaultDatabaseURL {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.BaseURL == "" {
		missing = append(missing, "BASE_URL")
	}
	if !isProductionSessionSecret(cfg.SessionSecret) {
		missing = append(missing, "SESSION_SECRET")
	}
	if cfg.AdminEmail == "" {
		missing = append(missing, "ADMIN_EMAIL")
	}
	if !isProductionAdminPassword(cfg.AdminPassword) {
		missing = append(missing, "ADMIN_PASSWORD")
	}

	if len(missing) > 0 {
		return fmt.Errorf("production configuration is missing or unsafe: %s", strings.Join(missing, ", "))
	}

	return nil
}

func isProductionSessionSecret(secret string) bool {
	secret = strings.TrimSpace(secret)
	if len(secret) < 32 {
		return false
	}

	return !strings.Contains(strings.ToLower(secret), "change-me")
}

func isProductionAdminPassword(password string) bool {
	password = strings.TrimSpace(password)
	if len(password) < 12 {
		return false
	}

	lower := strings.ToLower(password)
	return password != "admin123" && !strings.Contains(lower, "change-me")
}
