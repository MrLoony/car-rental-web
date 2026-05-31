package model

import (
	"testing"
	"time"
)

func TestBookingPrefillHoldsValues(t *testing.T) {
	expiresAt := time.Date(2026, 6, 1, 14, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 31, 14, 0, 0, 0, time.UTC)

	prefill := BookingPrefill{
		ID:        12,
		Token:     "prefill-token",
		Name:      "Stage Sixteen",
		Email:     "stage16@example.com",
		Phone:     "555-1601",
		PickupAt:  "2026-06-04T12:00",
		ReturnAt:  "2026-06-04T18:00",
		Message:   "Carry this booking form",
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
	}

	if prefill.ID != 12 {
		t.Fatalf("ID = %d, want 12", prefill.ID)
	}
	if prefill.Token != "prefill-token" {
		t.Fatalf("Token = %q, want %q", prefill.Token, "prefill-token")
	}
	if prefill.Name != "Stage Sixteen" {
		t.Fatalf("Name = %q, want %q", prefill.Name, "Stage Sixteen")
	}
	if prefill.Email != "stage16@example.com" {
		t.Fatalf("Email = %q, want %q", prefill.Email, "stage16@example.com")
	}
	if prefill.Phone != "555-1601" {
		t.Fatalf("Phone = %q, want %q", prefill.Phone, "555-1601")
	}
	if prefill.PickupAt != "2026-06-04T12:00" {
		t.Fatalf("PickupAt = %q, want %q", prefill.PickupAt, "2026-06-04T12:00")
	}
	if prefill.ReturnAt != "2026-06-04T18:00" {
		t.Fatalf("ReturnAt = %q, want %q", prefill.ReturnAt, "2026-06-04T18:00")
	}
	if prefill.Message != "Carry this booking form" {
		t.Fatalf("Message = %q, want %q", prefill.Message, "Carry this booking form")
	}
	if !prefill.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("ExpiresAt = %v, want %v", prefill.ExpiresAt, expiresAt)
	}
	if !prefill.CreatedAt.Equal(createdAt) {
		t.Fatalf("CreatedAt = %v, want %v", prefill.CreatedAt, createdAt)
	}
}
