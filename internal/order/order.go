package order

import (
	"errors"
	"math"
	"time"
)

var (
	ErrNotFound          = errors.New("order not found")
	ErrInvalidTransition = errors.New("transizione di stato non valida")
)

// Order status constants.
const (
	StatusPending   = "pending"
	StatusPrepared  = "prepared"
	StatusFulfilled = "fulfilled"
)

// Order is the domain representation of a prescription order.
type Order struct {
	ID                     int64
	PrescriptionID         int64
	CycleStartDate         time.Time
	EstimatedDepletionDate time.Time
	Status                 string
}

// CreateParams holds the data needed to create an order.
type CreateParams struct {
	PrescriptionID         int64
	CycleStartDate         time.Time
	EstimatedDepletionDate time.Time
}

// PrescriptionSummary is a lightweight prescription view used for order generation.
type PrescriptionSummary struct {
	ID               int64
	PatientID        int64
	UnitsPerBox      int
	DailyConsumption float64
	BoxStartDate     time.Time
}

// EstimatedDepletionDate calculates when this prescription's current box runs out.
func (p PrescriptionSummary) EstimatedDepletionDate() time.Time {
	days := math.Floor(float64(p.UnitsPerBox) / p.DailyConsumption)
	return p.BoxStartDate.AddDate(0, 0, int(days))
}

// DaysRemaining returns the number of days until depletion relative to now.
func (p PrescriptionSummary) DaysRemaining(now time.Time) int {
	depletion := p.EstimatedDepletionDate()
	now = now.Truncate(24 * time.Hour)
	depletion = depletion.Truncate(24 * time.Hour)
	return int(depletion.Sub(now).Hours() / 24)
}

// DashboardEntry combines order, prescription, and patient data for display.
type DashboardEntry struct {
	OrderID                int64
	PrescriptionID         int64
	CycleStartDate         time.Time
	EstimatedDepletionDate time.Time
	OrderStatus            string
	MedicationName         string
	UnitsPerBox            int
	DailyConsumption       float64
	BoxStartDate           time.Time
	PatientID              int64
	FirstName              string
	LastName               string
	Fulfillment            string
	DeliveryAddress        string
	Phone                  string
	Email                  string
}

// DaysRemaining returns the number of days until estimated depletion.
func (e DashboardEntry) DaysRemaining(now time.Time) int {
	now = now.Truncate(24 * time.Hour)
	depletion := e.EstimatedDepletionDate.Truncate(24 * time.Hour)
	return int(depletion.Sub(now).Hours() / 24)
}

// PrescriptionStatus classifies the entry: "ok" (>7), "approaching" (<=7), "depleted" (<=0).
func (e DashboardEntry) PrescriptionStatus(now time.Time) string {
	days := e.DaysRemaining(now)
	switch {
	case days <= 0:
		return "depleted"
	case days <= 7:
		return "approaching"
	default:
		return "ok"
	}
}

// NextStatus returns the next valid status in the lifecycle, or empty if terminal.
func NextStatus(current string) string {
	switch current {
	case StatusPending:
		return StatusPrepared
	case StatusPrepared:
		return StatusFulfilled
	default:
		return ""
	}
}
