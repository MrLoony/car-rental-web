package model

// VehicleSuggestion represents an alternative car for the requested rental period.
type VehicleSuggestion struct {
	Car            Car
	BillingDays    int
	EstimatedTotal float64
}
