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

func TestIsBlockingBookingStatus(t *testing.T) {
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
			want:   false,
		},
		{
			name:   "completed",
			status: BookingStatusCompleted,
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
			got := IsBlockingBookingStatus(tt.status)
			if got != tt.want {
				t.Fatalf("IsBlockingBookingStatus(%q) = %t, want %t", tt.status, got, tt.want)
			}
		})
	}
}

func TestBookingReturnBufferHours(t *testing.T) {
	if BookingReturnBufferHours != 4 {
		t.Fatalf("BookingReturnBufferHours = %d, want 4", BookingReturnBufferHours)
	}
}
