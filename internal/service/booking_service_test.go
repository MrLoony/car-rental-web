package service

import (
	"context"
	"testing"
	"time"

	"github.com/MrLoony/car-rental-web/internal/model"
)

func TestCalculateBillingDays(t *testing.T) {
	pickupAt := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		returnAt time.Time
		want     int
	}{
		{
			name:     "exactly 24 hours",
			returnAt: pickupAt.Add(24 * time.Hour),
			want:     1,
		},
		{
			name:     "25 hours",
			returnAt: pickupAt.Add(25 * time.Hour),
			want:     2,
		},
		{
			name:     "48 hours",
			returnAt: pickupAt.Add(48 * time.Hour),
			want:     2,
		},
		{
			name:     "49 hours",
			returnAt: pickupAt.Add(49 * time.Hour),
			want:     3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateBillingDays(pickupAt, tt.returnAt)
			if got != tt.want {
				t.Fatalf("calculateBillingDays() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCreateBookingInitializesFormErrors(t *testing.T) {
	service := BookingService{}

	id, form, err := service.CreateBooking(context.Background(), model.Car{}, model.BookingForm{})
	if err != nil {
		t.Fatalf("CreateBooking() error = %v", err)
	}

	if id != 0 {
		t.Fatalf("CreateBooking() id = %d, want 0", id)
	}

	if !form.HasErrors() {
		t.Fatal("CreateBooking() form has no validation errors")
	}

	if form.Errors["customer_name"] == "" {
		t.Fatal("CreateBooking() did not validate customer name")
	}
}
