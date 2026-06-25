package config

import (
	"strings"
	"testing"
)

func TestLoadDefaultsEmptyAppEnvToDevelopment(t *testing.T) {
	clearEmailEnv(t)
	t.Setenv("APP_ENV", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.AppEnv != appEnvDevelopment {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, appEnvDevelopment)
	}
	if cfg.IsProduction {
		t.Fatal("IsProduction = true, want false")
	}
}

func TestLoadDevelopmentAppEnv(t *testing.T) {
	clearEmailEnv(t)
	t.Setenv("APP_ENV", "development")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.AppEnv != appEnvDevelopment {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, appEnvDevelopment)
	}
	if cfg.IsProduction {
		t.Fatal("IsProduction = true, want false")
	}
}

func TestLoadProductionAppEnv(t *testing.T) {
	clearEmailEnv(t)
	setValidProductionEnv(t)
	t.Setenv("APP_ENV", "production")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.AppEnv != appEnvProduction {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, appEnvProduction)
	}
	if !cfg.IsProduction {
		t.Fatal("IsProduction = false, want true")
	}
}

func TestLoadUsesRenderPortWhenPresent(t *testing.T) {
	clearEmailEnv(t)
	t.Setenv("APP_PORT", "8081")
	t.Setenv("PORT", "10000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.AppPort != "10000" {
		t.Fatalf("AppPort = %q, want PORT value", cfg.AppPort)
	}
}

func TestLoadFallsBackToAppPortForDevelopment(t *testing.T) {
	clearEmailEnv(t)
	t.Setenv("APP_PORT", "8081")
	t.Setenv("PORT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.AppPort != "8081" {
		t.Fatalf("AppPort = %q, want APP_PORT value", cfg.AppPort)
	}
}

func TestLoadProductionRequiresSafeConfiguration(t *testing.T) {
	clearEmailEnv(t)
	t.Setenv("APP_ENV", "production")

	_, err := Load()
	assertConfigErrorContains(t, err, "DATABASE_URL")
	assertConfigErrorContains(t, err, "BASE_URL")
	assertConfigErrorContains(t, err, "SESSION_SECRET")
	assertConfigErrorContains(t, err, "ADMIN_EMAIL")
	assertConfigErrorContains(t, err, "ADMIN_PASSWORD")
}

func TestLoadProductionRejectsDemoAdminPassword(t *testing.T) {
	clearEmailEnv(t)
	setValidProductionEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("ADMIN_PASSWORD", "admin123")

	_, err := Load()
	assertConfigErrorContains(t, err, "ADMIN_PASSWORD")
}

func TestLoadUnknownAppEnvFallsBackToDevelopment(t *testing.T) {
	clearEmailEnv(t)
	t.Setenv("APP_ENV", "staging")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.AppEnv != appEnvDevelopment {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, appEnvDevelopment)
	}
	if cfg.IsProduction {
		t.Fatal("IsProduction = true, want false")
	}
}

func TestLoadDefaultEmailConfig(t *testing.T) {
	clearEmailEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.EmailEnabled {
		t.Fatal("EmailEnabled = true, want false")
	}
	if cfg.SMTPPort != 587 {
		t.Fatalf("SMTPPort = %d, want 587", cfg.SMTPPort)
	}
	if cfg.SMTPFromName != "Car Rental Web" {
		t.Fatalf("SMTPFromName = %q, want %q", cfg.SMTPFromName, "Car Rental Web")
	}
}

func TestLoadEmailDisabledAllowsEmptySMTPConfig(t *testing.T) {
	clearEmailEnv(t)
	t.Setenv("EMAIL_ENABLED", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.EmailEnabled {
		t.Fatal("EmailEnabled = true, want false")
	}
	if cfg.SMTPHost != "" {
		t.Fatalf("SMTPHost = %q, want empty", cfg.SMTPHost)
	}
	if cfg.SMTPFrom != "" {
		t.Fatalf("SMTPFrom = %q, want empty", cfg.SMTPFrom)
	}
	if cfg.AdminNotificationEmail != "" {
		t.Fatalf("AdminNotificationEmail = %q, want empty", cfg.AdminNotificationEmail)
	}
}

func TestLoadEmailEnabledRequiresSMTPHost(t *testing.T) {
	setValidEmailEnv(t)
	t.Setenv("SMTP_HOST", "")

	_, err := Load()
	assertConfigErrorContains(t, err, "SMTP_HOST")
}

func TestLoadEmailEnabledRequiresSMTPFrom(t *testing.T) {
	setValidEmailEnv(t)
	t.Setenv("SMTP_FROM", "")

	_, err := Load()
	assertConfigErrorContains(t, err, "SMTP_FROM")
}

func TestLoadEmailEnabledRequiresAdminNotificationEmail(t *testing.T) {
	setValidEmailEnv(t)
	t.Setenv("ADMIN_NOTIFICATION_EMAIL", "")

	_, err := Load()
	assertConfigErrorContains(t, err, "ADMIN_NOTIFICATION_EMAIL")
}

func TestLoadEmailEnabledWithRequiredConfig(t *testing.T) {
	setValidEmailEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if !cfg.EmailEnabled {
		t.Fatal("EmailEnabled = false, want true")
	}
	if cfg.SMTPPort != 587 {
		t.Fatalf("SMTPPort = %d, want 587", cfg.SMTPPort)
	}
}

func clearEmailEnv(t *testing.T) {
	t.Helper()

	t.Setenv("APP_PORT", "")
	t.Setenv("PORT", "")
	t.Setenv("BASE_URL", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SESSION_SECRET", "")
	t.Setenv("ADMIN_EMAIL", "")
	t.Setenv("ADMIN_PASSWORD", "")
	t.Setenv("EMAIL_ENABLED", "")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM", "")
	t.Setenv("SMTP_FROM_NAME", "")
	t.Setenv("ADMIN_NOTIFICATION_EMAIL", "")
}

func setValidProductionEnv(t *testing.T) {
	t.Helper()

	t.Setenv("BASE_URL", "https://car-rental.example.test")
	t.Setenv("DATABASE_URL", "postgres://prod_user:prod_password@db.example.test:5432/car_rental_web?sslmode=require")
	t.Setenv("SESSION_SECRET", "prod-session-secret-at-least-32-chars")
	t.Setenv("ADMIN_EMAIL", "admin@example.test")
	t.Setenv("ADMIN_PASSWORD", "replace-with-a-strong-password")
}

func setValidEmailEnv(t *testing.T) {
	t.Helper()

	t.Setenv("EMAIL_ENABLED", "true")
	t.Setenv("SMTP_HOST", "smtp.example.test")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM", "no-reply@example.test")
	t.Setenv("SMTP_FROM_NAME", "Car Rental Web")
	t.Setenv("ADMIN_NOTIFICATION_EMAIL", "admin@example.test")
}

func assertConfigErrorContains(t *testing.T, err error, expected string) {
	t.Helper()

	if err == nil {
		t.Fatalf("Load() error = nil, want error containing %q", expected)
	}
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("Load() error = %q, want to contain %q", err.Error(), expected)
	}
}
