package model

import "testing"

func TestVehicleSuggestionHoldsValues(t *testing.T) {
	car := Car{
		ID:          12,
		Brand:       "Toyota",
		Model:       "Camry",
		Slug:        "toyota-camry",
		PricePerDay: 65,
	}

	suggestion := VehicleSuggestion{
		Car:            car,
		BillingDays:    3,
		EstimatedTotal: 195,
	}

	if suggestion.Car.ID != car.ID {
		t.Fatalf("Car.ID = %d, want %d", suggestion.Car.ID, car.ID)
	}

	if suggestion.Car.Brand != car.Brand {
		t.Fatalf("Car.Brand = %q, want %q", suggestion.Car.Brand, car.Brand)
	}

	if suggestion.Car.Model != car.Model {
		t.Fatalf("Car.Model = %q, want %q", suggestion.Car.Model, car.Model)
	}

	if suggestion.BillingDays != 3 {
		t.Fatalf("BillingDays = %d, want 3", suggestion.BillingDays)
	}

	if suggestion.EstimatedTotal != 195 {
		t.Fatalf("EstimatedTotal = %f, want 195", suggestion.EstimatedTotal)
	}
}
