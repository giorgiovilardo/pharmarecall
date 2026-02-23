package order

import (
	"errors"
	"time"

	"github.com/giorgiovilardo/pharmarecall/internal/depletion"
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
	return depletion.EstimatedDate(p.UnitsPerBox, p.DailyConsumption, p.BoxStartDate)
}

// DaysRemaining returns the number of days until depletion relative to now.
func (p PrescriptionSummary) DaysRemaining(now time.Time) int {
	return depletion.DaysRemaining(p.EstimatedDepletionDate(), now)
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
	return depletion.DaysRemaining(e.EstimatedDepletionDate, now)
}

// PrescriptionStatus classifies the entry: "ok" (>7), "approaching" (<=7), "depleted" (<=0).
func (e DashboardEntry) PrescriptionStatus(now time.Time) string {
	return depletion.Status(e.DaysRemaining(now))
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
