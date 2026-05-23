package model

import "time"

// Booking represents a customer rental booking.
type Booking struct {
	ID    int64
	CarID int64

	CustomerName  string
	CustomerEmail string
	CustomerPhone string

	PickupAt time.Time
	ReturnAt time.Time

	BillingDays    int
	EstimatedTotal float64

	Message string
	Status  string

	CreatedAt time.Time
	UpdatedAt time.Time
}
