package model

import "testing"

func TestBookingFormSuggestedPickupAt(t *testing.T) {
	form := NewBookingForm()
	form.SuggestedPickupAt = "Jun 03, 2026 14:00"

	if form.SuggestedPickupAt != "Jun 03, 2026 14:00" {
		t.Fatalf("SuggestedPickupAt = %q", form.SuggestedPickupAt)
	}
}
