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

	t.Setenv("EMAIL_ENABLED", "")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM", "")
	t.Setenv("SMTP_FROM_NAME", "")
	t.Setenv("ADMIN_NOTIFICATION_EMAIL", "")
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
