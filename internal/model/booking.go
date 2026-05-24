package model

import "time"

const (
	BookingStatusPending   = "pending"
	BookingStatusConfirmed = "confirmed"
	BookingStatusCancelled = "cancelled"
	BookingStatusCompleted = "completed"

	BookingReturnBufferHours = 4
)

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

func IsValidBookingStatus(status string) bool {
	switch status {
	case BookingStatusPending, BookingStatusConfirmed, BookingStatusCancelled, BookingStatusCompleted:
		return true
	default:
		return false
	}
}

func IsBlockingBookingStatus(status string) bool {
	switch status {
	case BookingStatusPending, BookingStatusConfirmed:
		return true
	default:
		return false
	}
}
