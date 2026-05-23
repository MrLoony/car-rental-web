package model

import "testing"

func TestIsValidBookingStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "pending",
			status: BookingStatusPending,
			want:   true,
		},
		{
			name:   "confirmed",
			status: BookingStatusConfirmed,
			want:   true,
		},
		{
			name:   "cancelled",
			status: BookingStatusCancelled,
			want:   true,
		},
		{
			name:   "completed",
			status: BookingStatusCompleted,
			want:   true,
		},
		{
			name:   "empty",
			status: "",
			want:   false,
		},
		{
			name:   "unknown",
			status: "archived",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidBookingStatus(tt.status)
			if got != tt.want {
				t.Fatalf("IsValidBookingStatus(%q) = %t, want %t", tt.status, got, tt.want)
			}
		})
	}
}
