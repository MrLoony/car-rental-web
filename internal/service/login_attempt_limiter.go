package service

import (
	"strings"
	"sync"
	"time"
)

const (
	maxLoginFailedAttempts = 5
	loginLockoutDuration   = 10 * time.Minute
	loginAttemptWindow     = 15 * time.Minute
	emptyLoginAttemptKey   = "__empty__"
)

type LoginAttemptLimiter struct {
	mu       sync.Mutex
	attempts map[string]loginAttemptState
	now      func() time.Time
}

type loginAttemptState struct {
	FailedAttempts int
	LockedUntil    time.Time
	LastFailedAt   time.Time
}

func NewLoginAttemptLimiter() *LoginAttemptLimiter {
	return &LoginAttemptLimiter{
		attempts: make(map[string]loginAttemptState),
		now:      time.Now,
	}
}

func (l *LoginAttemptLimiter) IsLocked(email string) (bool, time.Duration) {
	key := normalizeLoginAttemptEmail(email)
	now := l.currentTime()

	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.attempts[key]
	if !ok {
		return false, 0
	}

	if state.LockedUntil.After(now) {
		return true, state.LockedUntil.Sub(now)
	}

	if !state.LockedUntil.IsZero() {
		delete(l.attempts, key)
	}

	return false, 0
}

func (l *LoginAttemptLimiter) RecordFailure(email string) {
	key := normalizeLoginAttemptEmail(email)
	now := l.currentTime()

	l.mu.Lock()
	defer l.mu.Unlock()

	state := l.attempts[key]
	if state.LastFailedAt.IsZero() || now.Sub(state.LastFailedAt) > loginAttemptWindow {
		state.FailedAttempts = 0
		state.LockedUntil = time.Time{}
	}

	state.FailedAttempts++
	state.LastFailedAt = now
	if state.FailedAttempts >= maxLoginFailedAttempts {
		state.LockedUntil = now.Add(loginLockoutDuration)
	}

	l.attempts[key] = state
}

func (l *LoginAttemptLimiter) RecordSuccess(email string) {
	key := normalizeLoginAttemptEmail(email)

	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.attempts, key)
}

func (l *LoginAttemptLimiter) currentTime() time.Time {
	if l.now == nil {
		return time.Now()
	}

	return l.now()
}

func normalizeLoginAttemptEmail(email string) string {
	normalized := strings.TrimSpace(strings.ToLower(email))
	if normalized == "" {
		return emptyLoginAttemptKey
	}

	return normalized
}
