package model

import "time"

// Car represents a rentable car.
type Car struct {
	ID           int64
	CategoryID   int64
	CategoryName string
	Brand        string
	Model        string
	Slug         string
	Year         int
	PricePerDay  float64
	Transmission string
	FuelType     string
	Seats        int
	IsAvailable  bool
	ArchivedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
