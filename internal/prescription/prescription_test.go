package prescription_test

import (
	"testing"
	"time"

	"github.com/giorgiovilardo/pharmarecall/internal/prescription"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestEstimatedDepletionDate(t *testing.T) {
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
		{
			name:             "60 units at 2/day",
			unitsPerBox:      60,
			dailyConsumption: 2,
			boxStartDate:     date(2026, 3, 1),
			want:             date(2026, 3, 31),
		},
		{
			name:             "fractional consumption 0.5/day",
			unitsPerBox:      15,
			dailyConsumption: 0.5,
			boxStartDate:     date(2026, 1, 1),
			want:             date(2026, 1, 31),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := prescription.Prescription{
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

func TestDaysRemaining(t *testing.T) {
	p := prescription.Prescription{
		UnitsPerBox:      30,
		DailyConsumption: 1,
		BoxStartDate:     date(2026, 1, 1),
	}
	// Depletion is Jan 31.

	tests := []struct {
		name string
		now  time.Time
		want int
	}{
		{"20 days before depletion", date(2026, 1, 11), 20},
		{"on depletion day", date(2026, 1, 31), 0},
		{"1 day past depletion", date(2026, 2, 1), -1},
		{"day after start", date(2026, 1, 2), 29},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.DaysRemaining(tt.now)
			if got != tt.want {
				t.Errorf("DaysRemaining() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestStatus(t *testing.T) {
	p := prescription.Prescription{
		UnitsPerBox:      30,
		DailyConsumption: 1,
		BoxStartDate:     date(2026, 1, 1),
	}
	// Depletion is Jan 31.

	tests := []struct {
		name string
		now  time.Time
		want string
	}{
		{"10 days remaining is ok", date(2026, 1, 21), prescription.StatusOk},
		{"8 days remaining is ok", date(2026, 1, 23), prescription.StatusOk},
		{"7 days remaining is approaching", date(2026, 1, 24), prescription.StatusApproaching},
		{"5 days remaining is approaching", date(2026, 1, 26), prescription.StatusApproaching},
		{"1 day remaining is approaching", date(2026, 1, 30), prescription.StatusApproaching},
		{"0 days remaining is depleted", date(2026, 1, 31), prescription.StatusDepleted},
		{"past depletion is depleted", date(2026, 2, 5), prescription.StatusDepleted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.Status(tt.now)
			if got != tt.want {
				t.Errorf("Status() = %q, want %q", got, tt.want)
			}
		})
	}
}
