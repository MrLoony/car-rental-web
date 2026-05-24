package model

import "time"

// AdminUser represents an administrator account.
type AdminUser struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
