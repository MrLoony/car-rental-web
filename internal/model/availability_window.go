package model

import "time"

// AvailabilityWindow represents a possible rental window for a car.
type AvailabilityWindow struct {
	StartAt        time.Time
	EndAt          time.Time
	BillingDays    int
	EstimatedTotal float64
}
