package patient

import "errors"

var (
	ErrNotFound = errors.New("patient not found")
)

// Fulfillment constants.
const (
	FulfillmentPickup   = "pickup"
	FulfillmentShipping = "shipping"
)

// Patient is the domain representation of a patient.
type Patient struct {
	ID              int64
	PharmacyID      int64
	FirstName       string
	LastName        string
	Phone           string
	Email           string
	DeliveryAddress string
	Fulfillment     string
	Notes           string
	Consensus       bool
	ConsensusDate   *string
}

// Summary is a patient list item.
type Summary struct {
	ID        int64
	FirstName string
	LastName  string
	Phone     string
	Email     string
	Consensus bool
}

// CreateParams holds the data needed to create a patient.
type CreateParams struct {
	PharmacyID      int64
	FirstName       string
	LastName        string
	Phone           string
	Email           string
	DeliveryAddress string
	Fulfillment     string
	Notes           string
}

// UpdateParams holds the data needed to update a patient.
type UpdateParams struct {
	ID              int64
	FirstName       string
	LastName        string
	Phone           string
	Email           string
	DeliveryAddress string
	Fulfillment     string
	Notes           string
}
