package model

import "testing"

func TestNormalizeDashboardRange(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want DashboardRange
	}{
		{name: "all", in: "all", want: DashboardRangeAll},
		{name: "last 30 days", in: "30d", want: DashboardRangeLast30Days},
		{name: "this month", in: "month", want: DashboardRangeThisMonth},
		{name: "empty defaults all", in: "", want: DashboardRangeAll},
		{name: "invalid defaults all", in: "invalid", want: DashboardRangeAll},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeDashboardRange(tt.in); got != tt.want {
				t.Fatalf("NormalizeDashboardRange(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
