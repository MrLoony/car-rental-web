package model

import "testing"

func TestNormalizeAdminCarAvailability(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "available",
			value: AdminCarAvailabilityAvailable,
			want:  AdminCarAvailabilityAvailable,
		},
		{
			name:  "unavailable",
			value: AdminCarAvailabilityUnavailable,
			want:  AdminCarAvailabilityUnavailable,
		},
		{
			name:  "trims whitespace",
			value: "  available  ",
			want:  AdminCarAvailabilityAvailable,
		},
		{
			name:  "normalizes uppercase",
			value: "UNAVAILABLE",
			want:  AdminCarAvailabilityUnavailable,
		},
		{
			name:  "empty falls back to all",
			value: "",
			want:  AdminCarAvailabilityAll,
		},
		{
			name:  "invalid falls back to all",
			value: "archived",
			want:  AdminCarAvailabilityAll,
		},
		{
			name:  "explicit all falls back to all",
			value: AdminCarAvailabilityAll,
			want:  AdminCarAvailabilityAll,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeAdminCarAvailability(tt.value)
			if got != tt.want {
				t.Fatalf("NormalizeAdminCarAvailability(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
