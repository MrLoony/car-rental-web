package config

import "testing"

func TestLoadDefaultsEmptyAppEnvToDevelopment(t *testing.T) {
	t.Setenv("APP_ENV", "")

	cfg := Load()

	if cfg.AppEnv != appEnvDevelopment {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, appEnvDevelopment)
	}
	if cfg.IsProduction {
		t.Fatal("IsProduction = true, want false")
	}
}

func TestLoadDevelopmentAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "development")

	cfg := Load()

	if cfg.AppEnv != appEnvDevelopment {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, appEnvDevelopment)
	}
	if cfg.IsProduction {
		t.Fatal("IsProduction = true, want false")
	}
}

func TestLoadProductionAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "production")

	cfg := Load()

	if cfg.AppEnv != appEnvProduction {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, appEnvProduction)
	}
	if !cfg.IsProduction {
		t.Fatal("IsProduction = false, want true")
	}
}

func TestLoadUnknownAppEnvFallsBackToDevelopment(t *testing.T) {
	t.Setenv("APP_ENV", "staging")

	cfg := Load()

	if cfg.AppEnv != appEnvDevelopment {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, appEnvDevelopment)
	}
	if cfg.IsProduction {
		t.Fatal("IsProduction = true, want false")
	}
}
