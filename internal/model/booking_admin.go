package model

import "time"

// BookingAdminView represents a booking joined with car data for admin pages.
type BookingAdminView struct {
	ID       int64
	CarID    int64
	CarBrand string
	CarModel string
	CarSlug  string

	CustomerName  string
	CustomerEmail string
	CustomerPhone string

	PickupAt       time.Time
	ReturnAt       time.Time
	BillingDays    int
	EstimatedTotal float64

	Message string
	Status  string

	CreatedAt time.Time
	UpdatedAt time.Time
}
