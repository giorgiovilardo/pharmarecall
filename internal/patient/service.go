package patient

import (
	"context"
	"errors"
	"fmt"
)

// ServiceDeps holds individual port interfaces — used by tests to inject only what's needed.
type ServiceDeps struct {
	Creator   PatientCreator
	Getter    PatientGetter
	Lister    PatientLister
	Updater   PatientUpdater
	Consensus ConsensusRecorder
}

// Service contains patient domain business logic.
type Service struct {
	deps ServiceDeps
}

// NewService is the production constructor — takes a Repository (satisfies all ports).
func NewService(repo Repository) *Service {
	return &Service{deps: ServiceDeps{
		Creator:   repo,
		Getter:    repo,
		Lister:    repo,
		Updater:   repo,
		Consensus: repo,
	}}
}

// NewServiceWith is the test constructor — inject only what you need, rest stays nil.
func NewServiceWith(d ServiceDeps) *Service {
	return &Service{deps: d}
}

// List returns all patients for a pharmacy.
func (s *Service) List(ctx context.Context, pharmacyID int64) ([]Summary, error) {
	return s.deps.Lister.List(ctx, pharmacyID)
}

// Get returns a patient by ID.
func (s *Service) Get(ctx context.Context, id int64) (Patient, error) {
	return s.deps.Getter.GetByID(ctx, id)
}

// HasConsensus returns whether the patient has recorded consensus.
func (s *Service) HasConsensus(ctx context.Context, patientID int64) (bool, error) {
	p, err := s.deps.Getter.GetByID(ctx, patientID)
	if err != nil {
		return false, fmt.Errorf("checking patient consensus: %w", err)
	}
	return p.Consensus, nil
}

// Create validates and creates a patient.
func (s *Service) Create(ctx context.Context, p CreateParams) (Patient, error) {
	if p.FirstName == "" || p.LastName == "" {
		return Patient{}, errors.New("il nome e il cognome sono obbligatori")
	}
	if p.Phone == "" && p.Email == "" {
		return Patient{}, errors.New("è necessario almeno un contatto (telefono o email)")
	}
	if p.Fulfillment == "" {
		p.Fulfillment = FulfillmentPickup
	}
	if p.Fulfillment == FulfillmentShipping && p.DeliveryAddress == "" {
		return Patient{}, errors.New("l'indirizzo di consegna è obbligatorio per la spedizione")
	}

	pt, err := s.deps.Creator.Create(ctx, p)
	if err != nil {
		return Patient{}, fmt.Errorf("creating patient: %w", err)
	}
	return pt, nil
}

// Update validates and updates a patient.
func (s *Service) Update(ctx context.Context, p UpdateParams) error {
	if p.FirstName == "" || p.LastName == "" {
		return errors.New("il nome e il cognome sono obbligatori")
	}
	if p.Phone == "" && p.Email == "" {
		return errors.New("è necessario almeno un contatto (telefono o email)")
	}
	if p.Fulfillment == FulfillmentShipping && p.DeliveryAddress == "" {
		return errors.New("l'indirizzo di consegna è obbligatorio per la spedizione")
	}

	if err := s.deps.Updater.Update(ctx, p); err != nil {
		return fmt.Errorf("updating patient: %w", err)
	}
	return nil
}

// SetConsensus records that a patient has given consensus.
func (s *Service) SetConsensus(ctx context.Context, id int64) error {
	if err := s.deps.Consensus.SetConsensus(ctx, id); err != nil {
		return fmt.Errorf("setting consensus: %w", err)
	}
	return nil
}
