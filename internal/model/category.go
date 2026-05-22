package model

import "time"

// Category represents a car category.
type Category struct {
	ID          int64
	Name        string
	Slug        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
