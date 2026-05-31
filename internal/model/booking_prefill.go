package model

import "time"

// BookingPrefill stores temporary booking form state for server-side prefill links.
type BookingPrefill struct {
	ID    int64
	Token string

	Name  string
	Email string
	Phone string

	PickupAt string
	ReturnAt string

	Message string

	ExpiresAt time.Time
	CreatedAt time.Time
}
