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

func TestFindAvailabilityWindowsNoBlockingBookings(t *testing.T) {
	requestedPickup := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	requestedReturn := requestedPickup.Add(48 * time.Hour)

	windows := findAvailabilityWindows(requestedPickup, requestedReturn, nil, 90)

	if len(windows) != 1 {
		t.Fatalf("len(windows) = %d, want 1", len(windows))
	}

	assertAvailabilityWindow(t, windows[0], requestedPickup, requestedReturn, 2, 180)
}

func TestFindAvailabilityWindowsSingleConflictReturnsNextWindowAfterBuffer(t *testing.T) {
	requestedPickup := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	requestedReturn := requestedPickup.Add(48 * time.Hour)
	blockingBookings := []model.Booking{
		{
			PickupAt: requestedPickup,
			ReturnAt: requestedReturn,
			Status:   model.BookingStatusPending,
		},
	}

	windows := findAvailabilityWindows(requestedPickup, requestedReturn, blockingBookings, 75)

	if len(windows) != 1 {
		t.Fatalf("len(windows) = %d, want 1", len(windows))
	}

	wantStart := requestedReturn.Add(time.Duration(model.BookingReturnBufferHours) * time.Hour)
	assertAvailabilityWindow(t, windows[0], wantStart, wantStart.Add(48*time.Hour), 2, 150)
}

func TestFindAvailabilityWindowsUsesLargeGapBetweenConflicts(t *testing.T) {
	requestedPickup := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	requestedReturn := requestedPickup.Add(24 * time.Hour)
	firstReturn := requestedPickup.Add(2 * time.Hour)
	secondPickup := firstReturn.Add(time.Duration(model.BookingReturnBufferHours)*time.Hour + 30*time.Hour)
	blockingBookings := []model.Booking{
		{
			PickupAt: requestedPickup,
			ReturnAt: firstReturn,
			Status:   model.BookingStatusPending,
		},
		{
			PickupAt: secondPickup,
			ReturnAt: secondPickup.Add(24 * time.Hour),
			Status:   model.BookingStatusConfirmed,
		},
	}

	windows := findAvailabilityWindows(requestedPickup, requestedReturn, blockingBookings, 120)

	if len(windows) != 2 {
		t.Fatalf("len(windows) = %d, want 2", len(windows))
	}

	wantFirstStart := firstReturn.Add(time.Duration(model.BookingReturnBufferHours) * time.Hour)
	assertAvailabilityWindow(t, windows[0], wantFirstStart, wantFirstStart.Add(24*time.Hour), 1, 120)

	wantSecondStart := secondPickup.Add(24*time.Hour + time.Duration(model.BookingReturnBufferHours)*time.Hour)
	assertAvailabilityWindow(t, windows[1], wantSecondStart, wantSecondStart.Add(24*time.Hour), 1, 120)
}

func TestFindAvailabilityWindowsSkipsGapTooSmall(t *testing.T) {
	requestedPickup := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	requestedReturn := requestedPickup.Add(24 * time.Hour)
	firstReturn := requestedPickup.Add(2 * time.Hour)
	secondPickup := firstReturn.Add(time.Duration(model.BookingReturnBufferHours)*time.Hour + 23*time.Hour)
	blockingBookings := []model.Booking{
		{
			PickupAt: requestedPickup,
			ReturnAt: firstReturn,
			Status:   model.BookingStatusPending,
		},
		{
			PickupAt: secondPickup,
			ReturnAt: secondPickup.Add(24 * time.Hour),
			Status:   model.BookingStatusConfirmed,
		},
	}

	windows := findAvailabilityWindows(requestedPickup, requestedReturn, blockingBookings, 80)

	if len(windows) != 1 {
		t.Fatalf("len(windows) = %d, want 1", len(windows))
	}

	wantStart := secondPickup.Add(24*time.Hour + time.Duration(model.BookingReturnBufferHours)*time.Hour)
	assertAvailabilityWindow(t, windows[0], wantStart, wantStart.Add(24*time.Hour), 1, 80)
}

func TestFindAvailabilityWindowsLimitsSuggestionsToThree(t *testing.T) {
	requestedPickup := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	requestedReturn := requestedPickup.Add(24 * time.Hour)
	blockingBookings := []model.Booking{
		{
			PickupAt: requestedPickup,
			ReturnAt: requestedPickup.Add(2 * time.Hour),
			Status:   model.BookingStatusPending,
		},
		{
			PickupAt: requestedPickup.Add(48 * time.Hour),
			ReturnAt: requestedPickup.Add(50 * time.Hour),
			Status:   model.BookingStatusPending,
		},
		{
			PickupAt: requestedPickup.Add(96 * time.Hour),
			ReturnAt: requestedPickup.Add(98 * time.Hour),
			Status:   model.BookingStatusPending,
		},
	}

	windows := findAvailabilityWindows(requestedPickup, requestedReturn, blockingBookings, 60)

	if len(windows) != 3 {
		t.Fatalf("len(windows) = %d, want 3", len(windows))
	}
}

func TestFindAvailabilityWindowsCalculatesCeilBillingDaysAndTotal(t *testing.T) {
	requestedPickup := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	requestedReturn := requestedPickup.Add(25 * time.Hour)

	windows := findAvailabilityWindows(requestedPickup, requestedReturn, nil, 45)

	if len(windows) != 1 {
		t.Fatalf("len(windows) = %d, want 1", len(windows))
	}

	assertAvailabilityWindow(t, windows[0], requestedPickup, requestedReturn, 2, 90)
}

func TestAlternativeVehiclePriceRange(t *testing.T) {
	minPrice, maxPrice := alternativeVehiclePriceRange(100)

	if minPrice != 80 {
		t.Fatalf("minPrice = %f, want 80", minPrice)
	}

	if maxPrice != 120 {
		t.Fatalf("maxPrice = %f, want 120", maxPrice)
	}
}

func TestBuildVehicleSuggestionsEmptyCars(t *testing.T) {
	pickupAt := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	returnAt := pickupAt.Add(24 * time.Hour)

	suggestions := buildVehicleSuggestions(nil, pickupAt, returnAt)

	if suggestions != nil {
		t.Fatalf("suggestions = %#v, want nil", suggestions)
	}
}

func TestBuildVehicleSuggestionsCalculatesBillingAndTotals(t *testing.T) {
	pickupAt := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	returnAt := pickupAt.Add(25 * time.Hour)
	cars := []model.Car{
		{
			ID:          10,
			Brand:       "Hyundai",
			Model:       "Elantra",
			PricePerDay: 50,
		},
		{
			ID:          11,
			Brand:       "Toyota",
			Model:       "Camry",
			PricePerDay: 65,
		},
	}

	suggestions := buildVehicleSuggestions(cars, pickupAt, returnAt)

	if len(suggestions) != 2 {
		t.Fatalf("len(suggestions) = %d, want 2", len(suggestions))
	}

	if suggestions[0].Car.ID != 10 {
		t.Fatalf("suggestions[0].Car.ID = %d, want 10", suggestions[0].Car.ID)
	}

	if suggestions[0].BillingDays != 2 {
		t.Fatalf("suggestions[0].BillingDays = %d, want 2", suggestions[0].BillingDays)
	}

	if suggestions[0].EstimatedTotal != 100 {
		t.Fatalf("suggestions[0].EstimatedTotal = %f, want 100", suggestions[0].EstimatedTotal)
	}

	if suggestions[1].BillingDays != 2 {
		t.Fatalf("suggestions[1].BillingDays = %d, want 2", suggestions[1].BillingDays)
	}

	if suggestions[1].EstimatedTotal != 130 {
		t.Fatalf("suggestions[1].EstimatedTotal = %f, want 130", suggestions[1].EstimatedTotal)
	}
}

func assertAvailabilityWindow(t *testing.T, window model.AvailabilityWindow, startAt time.Time, endAt time.Time, billingDays int, estimatedTotal float64) {
	t.Helper()

	if !window.StartAt.Equal(startAt) {
		t.Fatalf("StartAt = %v, want %v", window.StartAt, startAt)
	}

	if !window.EndAt.Equal(endAt) {
		t.Fatalf("EndAt = %v, want %v", window.EndAt, endAt)
	}

	if window.BillingDays != billingDays {
		t.Fatalf("BillingDays = %d, want %d", window.BillingDays, billingDays)
	}

	if window.EstimatedTotal != estimatedTotal {
		t.Fatalf("EstimatedTotal = %f, want %f", window.EstimatedTotal, estimatedTotal)
	}
}
