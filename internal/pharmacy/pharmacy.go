package pharmacy

import "errors"

var (
	ErrNotFound       = errors.New("pharmacy not found")
	ErrDuplicateEmail = errors.New("email already in use")
)

// Pharmacy is the domain representation of a pharmacy.
type Pharmacy struct {
	ID      int64
	Name    string
	Address string
	Phone   string
	Email   string
}

// Summary is a pharmacy list item with personnel count.
type Summary struct {
	ID             int64
	Name           string
	Address        string
	PersonnelCount int64
}

// PersonnelMember is a user belonging to a pharmacy.
type PersonnelMember struct {
	ID    int64
	Name  string
	Email string
	Role  string
}

// CreateParams holds the data needed to create a pharmacy with its owner.
type CreateParams struct {
	Name          string
	Address       string
	Phone         string
	Email         string
	OwnerName     string
	OwnerEmail    string
	OwnerPassword string
}

// UpdateParams holds the data needed to update a pharmacy.
type UpdateParams struct {
	ID      int64
	Name    string
	Address string
	Phone   string
	Email   string
}

// CreatePersonnelParams holds the data needed to create a personnel member.
type CreatePersonnelParams struct {
	PharmacyID int64
	Name       string
	Email      string
	Password   string
	Role       string
}
