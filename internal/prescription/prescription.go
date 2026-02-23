package prescription

import (
	"errors"
	"math"
	"time"
)

var (
	ErrNotFound    = errors.New("prescription not found")
	ErrNoConsensus = errors.New("il paziente deve dare il consenso prima di aggiungere prescrizioni")
)

// Status constants for prescription depletion classification.
const (
	StatusOk          = "ok"
	StatusApproaching = "approaching"
	StatusDepleted    = "depleted"
)

// Prescription is the domain representation of a recurring prescription.
type Prescription struct {
	ID               int64
	PatientID        int64
	MedicationName   string
	UnitsPerBox      int
	DailyConsumption float64
	BoxStartDate     time.Time
}

// EstimatedDepletionDate returns the date when the current box is expected to run out.
// Formula: box_start_date + floor(units_per_box / daily_consumption) days.
func (p Prescription) EstimatedDepletionDate() time.Time {
	days := math.Floor(float64(p.UnitsPerBox) / p.DailyConsumption)
	return p.BoxStartDate.AddDate(0, 0, int(days))
}

// DaysRemaining returns the number of days until depletion relative to the given date.
// Negative values mean the prescription is past depletion.
func (p Prescription) DaysRemaining(now time.Time) int {
	depletion := p.EstimatedDepletionDate()
	// Truncate both to date-only to avoid time-of-day affecting the calculation.
	now = now.Truncate(24 * time.Hour)
	depletion = depletion.Truncate(24 * time.Hour)
	return int(depletion.Sub(now).Hours() / 24)
}

// Status classifies the prescription based on days remaining:
// "ok" (>7 days), "approaching" (<=7 days), "depleted" (<=0 days).
func (p Prescription) Status(now time.Time) string {
	days := p.DaysRemaining(now)
	switch {
	case days <= 0:
		return StatusDepleted
	case days <= 7:
		return StatusApproaching
	default:
		return StatusOk
	}
}

// CreateParams holds the data needed to create a prescription.
type CreateParams struct {
	PatientID        int64
	MedicationName   string
	UnitsPerBox      int
	DailyConsumption float64
	BoxStartDate     time.Time
}

// UpdateParams holds the data needed to update a prescription.
type UpdateParams struct {
	ID               int64
	MedicationName   string
	UnitsPerBox      int
	DailyConsumption float64
	BoxStartDate     time.Time
}

// RefillParams holds the data needed to record a refill.
type RefillParams struct {
	PrescriptionID int64
	NewStartDate   time.Time
}
