package model

import "testing"

func TestNormalizeAdminBookingStatus(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "all",
			value: AdminBookingStatusAll,
			want:  AdminBookingStatusAll,
		},
		{
			name:  "pending",
			value: BookingStatusPending,
			want:  BookingStatusPending,
		},
		{
			name:  "confirmed",
			value: BookingStatusConfirmed,
			want:  BookingStatusConfirmed,
		},
		{
			name:  "cancelled",
			value: BookingStatusCancelled,
			want:  BookingStatusCancelled,
		},
		{
			name:  "completed",
			value: BookingStatusCompleted,
			want:  BookingStatusCompleted,
		},
		{
			name:  "trims whitespace",
			value: "  pending  ",
			want:  BookingStatusPending,
		},
		{
			name:  "normalizes uppercase",
			value: "CONFIRMED",
			want:  BookingStatusConfirmed,
		},
		{
			name:  "empty falls back to all",
			value: "",
			want:  AdminBookingStatusAll,
		},
		{
			name:  "invalid falls back to all",
			value: "review",
			want:  AdminBookingStatusAll,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeAdminBookingStatus(tt.value)
			if got != tt.want {
				t.Fatalf("NormalizeAdminBookingStatus(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
