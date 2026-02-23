package prescription

import (
	"errors"
	"time"

	"github.com/giorgiovilardo/pharmarecall/internal/depletion"
)

var (
	ErrNotFound              = errors.New("prescription not found")
	ErrNoConsensus           = errors.New("il paziente deve dare il consenso prima di aggiungere prescrizioni")
	ErrMedicationRequired    = errors.New("il nome del farmaco è obbligatorio")
	ErrInvalidUnitsPerBox    = errors.New("le unità per confezione devono essere maggiori di zero")
	ErrInvalidConsumption    = errors.New("il consumo giornaliero deve essere maggiore di zero")
	ErrStartDateRequired     = errors.New("la data di inizio confezione è obbligatoria")
	ErrConsumptionExceedsBox = errors.New("il consumo giornaliero deve essere inferiore alle unità per confezione (la confezione deve durare almeno un giorno)")
)

// Status constants — re-exported from depletion for backward compatibility.
const (
	StatusOk          = depletion.StatusOk
	StatusApproaching = depletion.StatusApproaching
	StatusDepleted    = depletion.StatusDepleted
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
func (p Prescription) EstimatedDepletionDate() time.Time {
	return depletion.EstimatedDate(p.UnitsPerBox, p.DailyConsumption, p.BoxStartDate)
}

// DaysRemaining returns the number of days until depletion relative to the given date.
func (p Prescription) DaysRemaining(now time.Time) int {
	return depletion.DaysRemaining(p.EstimatedDepletionDate(), now)
}

// Status classifies the prescription based on days remaining.
func (p Prescription) Status(now time.Time) string {
	return depletion.Status(p.DaysRemaining(now))
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
