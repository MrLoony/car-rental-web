package model

import (
	"testing"
	"time"
)

func TestAvailabilityWindowHoldsValues(t *testing.T) {
	startAt := time.Date(2026, time.June, 3, 14, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, time.June, 5, 14, 0, 0, 0, time.UTC)

	window := AvailabilityWindow{
		StartAt:        startAt,
		EndAt:          endAt,
		BillingDays:    2,
		EstimatedTotal: 180,
	}

	if !window.StartAt.Equal(startAt) {
		t.Fatalf("StartAt = %v, want %v", window.StartAt, startAt)
	}

	if !window.EndAt.Equal(endAt) {
		t.Fatalf("EndAt = %v, want %v", window.EndAt, endAt)
	}

	if window.BillingDays != 2 {
		t.Fatalf("BillingDays = %d, want 2", window.BillingDays)
	}

	if window.EstimatedTotal != 180 {
		t.Fatalf("EstimatedTotal = %f, want 180", window.EstimatedTotal)
	}
}
