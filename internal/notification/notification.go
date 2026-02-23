package notification

import (
	"errors"
	"math"
	"time"
)

var (
	ErrNotFound = errors.New("notification not found")
)

// Transition type constants.
const (
	TransitionApproaching = "approaching"
)

// Notification is the domain representation of a personnel notification.
type Notification struct {
	ID               int64
	PharmacyID       int64
	PrescriptionID   int64
	TransitionType   string
	Read             bool
	CreatedAt        time.Time
	MedicationName   string
	UnitsPerBox      int
	DailyConsumption float64
	BoxStartDate     time.Time
	PatientID        int64
	FirstName        string
	LastName         string
}

// EstimatedDepletionDate calculates when the prescription's current box runs out.
func (n Notification) EstimatedDepletionDate() time.Time {
	days := math.Floor(float64(n.UnitsPerBox) / n.DailyConsumption)
	return n.BoxStartDate.AddDate(0, 0, int(days))
}
