package repository

import (
	"strings"
	"testing"

	"github.com/MrLoony/car-rental-web/internal/model"
)

func TestAppendDashboardRangeFilter(t *testing.T) {
	tests := []struct {
		name        string
		rangeValue  model.DashboardRange
		want        string
		wantMissing string
	}{
		{
			name:        "all time",
			rangeValue:  model.DashboardRangeAll,
			wantMissing: "created_at >=",
		},
		{
			name:       "last 30 days",
			rangeValue: model.DashboardRangeLast30Days,
			want:       "AND b.created_at >= NOW() - interval '30 days'",
		},
		{
			name:       "this month",
			rangeValue: model.DashboardRangeThisMonth,
			want:       "AND b.created_at >= date_trunc('month', NOW()) AND b.created_at < date_trunc('month', NOW()) + interval '1 month'",
		},
		{
			name:        "invalid defaults all",
			rangeValue:  model.DashboardRange("invalid"),
			wantMissing: "created_at >=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var query strings.Builder
			query.WriteString("WHERE TRUE")
			appendDashboardRangeFilter(&query, "b.created_at", tt.rangeValue)
			got := query.String()

			if tt.want != "" && !strings.Contains(got, tt.want) {
				t.Fatalf("query = %q, want to contain %q", got, tt.want)
			}
			if tt.wantMissing != "" && strings.Contains(got, tt.wantMissing) {
				t.Fatalf("query = %q, did not want to contain %q", got, tt.wantMissing)
			}
		})
	}
}
