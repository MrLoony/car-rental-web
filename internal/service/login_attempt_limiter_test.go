package service

import (
	"testing"
	"time"
)

func TestLoginAttemptLimiterLocksAfterFailures(t *testing.T) {
	now := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	limiter := newTestLoginAttemptLimiter(&now)

	for i := 0; i < maxLoginFailedAttempts-1; i++ {
		limiter.RecordFailure("admin@example.com")
	}

	if locked, _ := limiter.IsLocked("admin@example.com"); locked {
		t.Fatal("IsLocked() = true before max failures")
	}

	limiter.RecordFailure("admin@example.com")
	locked, remaining := limiter.IsLocked("admin@example.com")
	if !locked {
		t.Fatal("IsLocked() = false after max failures")
	}
	if remaining <= 0 {
		t.Fatalf("remaining = %s, want positive duration", remaining)
	}
}

func TestLoginAttemptLimiterNormalizesEmail(t *testing.T) {
	now := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	limiter := newTestLoginAttemptLimiter(&now)

	for i := 0; i < maxLoginFailedAttempts; i++ {
		limiter.RecordFailure(" Admin@Example.COM ")
	}

	if locked, _ := limiter.IsLocked("admin@example.com"); !locked {
		t.Fatal("IsLocked() = false, want true for normalized email")
	}
}

func TestLoginAttemptLimiterRecordSuccessClearsFailures(t *testing.T) {
	now := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	limiter := newTestLoginAttemptLimiter(&now)

	for i := 0; i < maxLoginFailedAttempts; i++ {
		limiter.RecordFailure("admin@example.com")
	}
	limiter.RecordSuccess("admin@example.com")

	if locked, _ := limiter.IsLocked("admin@example.com"); locked {
		t.Fatal("IsLocked() = true after successful login")
	}

	for i := 0; i < maxLoginFailedAttempts-1; i++ {
		limiter.RecordFailure("admin@example.com")
	}
	if locked, _ := limiter.IsLocked("admin@example.com"); locked {
		t.Fatal("IsLocked() = true before max failures after success reset")
	}
}

func TestLoginAttemptLimiterResetsAfterAttemptWindow(t *testing.T) {
	now := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	limiter := newTestLoginAttemptLimiter(&now)

	for i := 0; i < maxLoginFailedAttempts-1; i++ {
		limiter.RecordFailure("admin@example.com")
	}

	now = now.Add(loginAttemptWindow + time.Second)
	for i := 0; i < maxLoginFailedAttempts-1; i++ {
		limiter.RecordFailure("admin@example.com")
	}

	if locked, _ := limiter.IsLocked("admin@example.com"); locked {
		t.Fatal("IsLocked() = true, want false after old failures fall outside window")
	}
}

func TestLoginAttemptLimiterClearsExpiredLock(t *testing.T) {
	now := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	limiter := newTestLoginAttemptLimiter(&now)

	for i := 0; i < maxLoginFailedAttempts; i++ {
		limiter.RecordFailure("admin@example.com")
	}

	now = now.Add(loginLockoutDuration + time.Second)
	if locked, _ := limiter.IsLocked("admin@example.com"); locked {
		t.Fatal("IsLocked() = true after lockout expired")
	}
}

func newTestLoginAttemptLimiter(now *time.Time) *LoginAttemptLimiter {
	limiter := NewLoginAttemptLimiter()
	limiter.now = func() time.Time {
		return *now
	}

	return limiter
}
