package pharmacy

import "context"

// PharmacyCreator creates a pharmacy and its owner in a single transaction.
type PharmacyCreator interface {
	CreateWithOwner(ctx context.Context, p CreateParams, ownerPasswordHash string) (Pharmacy, error)
}

// PharmacyGetter fetches a pharmacy by ID.
type PharmacyGetter interface {
	GetByID(ctx context.Context, id int64) (Pharmacy, error)
}

// PharmacyLister lists all pharmacies with personnel counts.
type PharmacyLister interface {
	List(ctx context.Context) ([]Summary, error)
}

// PharmacyUpdater updates a pharmacy in a transaction.
type PharmacyUpdater interface {
	Update(ctx context.Context, p UpdateParams) error
}

// PersonnelLister lists personnel for a pharmacy.
type PersonnelLister interface {
	ListPersonnel(ctx context.Context, pharmacyID int64) ([]PersonnelMember, error)
}

// PersonnelCreator creates a personnel member in a transaction.
type PersonnelCreator interface {
	CreatePersonnel(ctx context.Context, p CreatePersonnelParams, passwordHash string) (PersonnelMember, error)
}

// Repository composes all ports — used only by NewService for convenient wiring.
type Repository interface {
	PharmacyCreator
	PharmacyGetter
	PharmacyLister
	PharmacyUpdater
	PersonnelLister
	PersonnelCreator
}
