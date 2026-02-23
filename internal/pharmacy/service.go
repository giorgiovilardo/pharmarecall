package pharmacy

import (
	"context"
	"fmt"
)

// ServiceDeps holds individual port interfaces — used by tests to inject only what's needed.
type ServiceDeps struct {
	Creator     PharmacyCreator
	Getter      PharmacyGetter
	Lister      PharmacyLister
	Updater     PharmacyUpdater
	Personnel   PersonnelLister
	PersCreator PersonnelCreator
	Hasher      func(string) (string, error)
}

// Service contains pharmacy domain business logic.
type Service struct {
	deps ServiceDeps
}

// NewService is the production constructor — takes a Repository (satisfies all ports).
func NewService(repo Repository, hasher func(string) (string, error)) *Service {
	return &Service{deps: ServiceDeps{
		Creator:     repo,
		Getter:      repo,
		Lister:      repo,
		Updater:     repo,
		Personnel:   repo,
		PersCreator: repo,
		Hasher:      hasher,
	}}
}

// NewServiceWith is the test constructor — inject only what you need, rest stays nil.
func NewServiceWith(d ServiceDeps) *Service {
	return &Service{deps: d}
}

// CreateWithOwner validates, hashes the owner password, and creates a pharmacy with its owner.
func (s *Service) CreateWithOwner(ctx context.Context, p CreateParams) (Pharmacy, error) {
	hash, err := s.deps.Hasher(p.OwnerPassword)
	if err != nil {
		return Pharmacy{}, fmt.Errorf("hashing owner password: %w", err)
	}

	ph, err := s.deps.Creator.CreateWithOwner(ctx, p, hash)
	if err != nil {
		return Pharmacy{}, err
	}

	return ph, nil
}

// List returns all pharmacies with personnel counts.
func (s *Service) List(ctx context.Context) ([]Summary, error) {
	return s.deps.Lister.List(ctx)
}

// Get returns a pharmacy by ID.
func (s *Service) Get(ctx context.Context, id int64) (Pharmacy, error) {
	return s.deps.Getter.GetByID(ctx, id)
}

// Update updates a pharmacy.
func (s *Service) Update(ctx context.Context, p UpdateParams) error {
	return s.deps.Updater.Update(ctx, p)
}

// ListPersonnel returns personnel for a pharmacy.
func (s *Service) ListPersonnel(ctx context.Context, pharmacyID int64) ([]PersonnelMember, error) {
	return s.deps.Personnel.ListPersonnel(ctx, pharmacyID)
}

// CreatePersonnel validates, hashes the password, and creates a personnel member.
func (s *Service) CreatePersonnel(ctx context.Context, p CreatePersonnelParams) (PersonnelMember, error) {
	hash, err := s.deps.Hasher(p.Password)
	if err != nil {
		return PersonnelMember{}, fmt.Errorf("hashing personnel password: %w", err)
	}

	m, err := s.deps.PersCreator.CreatePersonnel(ctx, p, hash)
	if err != nil {
		return PersonnelMember{}, err
	}

	return m, nil
}
