package model

// BookingForm holds raw booking form input.
type BookingForm struct {
	CustomerName  string
	CustomerEmail string
	CustomerPhone string

	PickupAt string
	ReturnAt string

	Message                      string
	SuggestedPickupAt            string
	SuggestedAvailabilityWindows []AvailabilityWindow
	SuggestedVehicles            []VehicleSuggestion

	Errors map[string]string
}

func NewBookingForm() BookingForm {
	return BookingForm{
		Errors: make(map[string]string),
	}
}

func (f BookingForm) HasErrors() bool {
	return len(f.Errors) > 0
}
