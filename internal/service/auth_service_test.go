package service

import (
	"context"
	"testing"
	"time"

	"github.com/MrLoony/car-rental-web/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func TestValidateLoginForm(t *testing.T) {
	form := normalizeLoginForm(model.LoginForm{})
	validateLoginForm(&form)

	if form.Errors["email"] != "Enter your email address." {
		t.Fatalf("email error = %q", form.Errors["email"])
	}

	if form.Errors["password"] != "Enter your password." {
		t.Fatalf("password error = %q", form.Errors["password"])
	}
}

func TestPasswordMatchesHashWrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	if passwordMatchesHash("wrong-password", string(hash)) {
		t.Fatal("passwordMatchesHash() = true for wrong password")
	}
}

func TestPasswordMatchesHashCorrectPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	if !passwordMatchesHash("admin123", string(hash)) {
		t.Fatal("passwordMatchesHash() = false for correct password")
	}
}

func TestAuthenticateLockedEmailReturnsGenericError(t *testing.T) {
	now := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	limiter := newTestLoginAttemptLimiter(&now)
	for i := 0; i < maxLoginFailedAttempts; i++ {
		limiter.RecordFailure("admin@example.com")
	}

	service := NewAuthService(nil, limiter)
	_, form, err := service.Authenticate(context.Background(), model.LoginForm{
		Email:    "admin@example.com",
		Password: "admin123",
	})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if form.Errors["credentials"] != invalidCredentialsLaterMessage {
		t.Fatalf("credentials error = %q, want %q", form.Errors["credentials"], invalidCredentialsLaterMessage)
	}
}

func TestAuthenticateValidationErrorsDoNotRecordFailure(t *testing.T) {
	now := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	limiter := newTestLoginAttemptLimiter(&now)
	service := NewAuthService(nil, limiter)

	_, form, err := service.Authenticate(context.Background(), model.LoginForm{})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if form.Errors["email"] == "" || form.Errors["password"] == "" {
		t.Fatal("Authenticate() did not return field validation errors")
	}

	for i := 0; i < maxLoginFailedAttempts-1; i++ {
		limiter.RecordFailure("")
	}
	if locked, _ := limiter.IsLocked(""); locked {
		t.Fatal("empty email was locked too early; validation likely recorded a failure")
	}
}
