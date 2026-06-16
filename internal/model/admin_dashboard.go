package model

import "time"

// BookingStats contains simple booking counts for the admin dashboard.
type BookingStats struct {
	Total     int
	Pending   int
	Confirmed int
	Cancelled int
	Completed int
}

// RevenueStats contains simple revenue totals for the admin dashboard.
type RevenueStats struct {
	TotalPotential float64
	Pending        float64
	Confirmed      float64
	Completed      float64
	Cancelled      float64
}

// RecentBookingActivity contains the compact booking data shown on the admin dashboard.
type RecentBookingActivity struct {
	ID           int64
	CustomerName string
	CarName      string
	Status       string
	PickupTime   time.Time
	ReturnTime   time.Time
	CreatedAt    time.Time
}
