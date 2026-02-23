package order_test

import (
	"testing"
	"time"

	"github.com/giorgiovilardo/pharmarecall/internal/order"
)

func TestPrescriptionSummaryEstimatedDepletionDate(t *testing.T) {
	tests := []struct {
		name             string
		unitsPerBox      int
		dailyConsumption float64
		boxStartDate     time.Time
		want             time.Time
	}{
		{
			name:             "30 units at 1/day",
			unitsPerBox:      30,
			dailyConsumption: 1,
			boxStartDate:     date(2026, 1, 1),
			want:             date(2026, 1, 31),
		},
		{
			name:             "100 units at 3/day floors to 33 days",
			unitsPerBox:      100,
			dailyConsumption: 3,
			boxStartDate:     date(2026, 1, 1),
			want:             date(2026, 2, 3),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := order.PrescriptionSummary{
				UnitsPerBox:      tt.unitsPerBox,
				DailyConsumption: tt.dailyConsumption,
				BoxStartDate:     tt.boxStartDate,
			}
			got := p.EstimatedDepletionDate()
			if !got.Equal(tt.want) {
				t.Errorf("EstimatedDepletionDate() = %s, want %s", got.Format("2006-01-02"), tt.want.Format("2006-01-02"))
			}
		})
	}
}

func TestDashboardEntryPrescriptionStatus(t *testing.T) {
	tests := []struct {
		name          string
		depletionDate time.Time
		now           time.Time
		wantStatus    string
		wantRemaining int
	}{
		{"ok - 10 days remaining", date(2026, 2, 10), date(2026, 1, 31), "ok", 10},
		{"approaching - 5 days remaining", date(2026, 2, 5), date(2026, 1, 31), "approaching", 5},
		{"approaching - 7 days remaining", date(2026, 2, 7), date(2026, 1, 31), "approaching", 7},
		{"depleted - 0 days remaining", date(2026, 1, 31), date(2026, 1, 31), "depleted", 0},
		{"depleted - past", date(2026, 1, 25), date(2026, 1, 31), "depleted", -6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := order.DashboardEntry{EstimatedDepletionDate: tt.depletionDate}
			gotStatus := e.PrescriptionStatus(tt.now)
			gotRemaining := e.DaysRemaining(tt.now)
			if gotStatus != tt.wantStatus {
				t.Errorf("PrescriptionStatus() = %q, want %q", gotStatus, tt.wantStatus)
			}
			if gotRemaining != tt.wantRemaining {
				t.Errorf("DaysRemaining() = %d, want %d", gotRemaining, tt.wantRemaining)
			}
		})
	}
}

func TestNextStatus(t *testing.T) {
	tests := []struct {
		current string
		want    string
	}{
		{order.StatusPending, order.StatusPrepared},
		{order.StatusPrepared, order.StatusFulfilled},
		{order.StatusFulfilled, ""},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.current, func(t *testing.T) {
			got := order.NextStatus(tt.current)
			if got != tt.want {
				t.Errorf("NextStatus(%q) = %q, want %q", tt.current, got, tt.want)
			}
		})
	}
}
