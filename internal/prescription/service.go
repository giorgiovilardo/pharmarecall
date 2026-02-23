package prescription

import (
	"context"
	"errors"
	"fmt"
)

// ConsensusChecker checks if a patient has given consensus.
type ConsensusChecker interface {
	HasConsensus(ctx context.Context, patientID int64) (bool, error)
}

// ServiceDeps holds individual port interfaces — used by tests to inject only what's needed.
type ServiceDeps struct {
	Creator   PrescriptionCreator
	Getter    PrescriptionGetter
	Lister    PrescriptionLister
	Updater   PrescriptionUpdater
	Refill    RefillRecorder
	Consensus ConsensusChecker
}

// Service contains prescription domain business logic.
type Service struct {
	deps ServiceDeps
}

// NewService is the production constructor — takes a Repository (satisfies all ports).
func NewService(repo Repository, consensus ConsensusChecker) *Service {
	return &Service{deps: ServiceDeps{
		Creator:   repo,
		Getter:    repo,
		Lister:    repo,
		Updater:   repo,
		Refill:    repo,
		Consensus: consensus,
	}}
}

// NewServiceWith is the test constructor — inject only what you need, rest stays nil.
func NewServiceWith(d ServiceDeps) *Service {
	return &Service{deps: d}
}

// Create validates and creates a prescription. Blocks if the patient has no consensus.
func (s *Service) Create(ctx context.Context, p CreateParams) (Prescription, error) {
	if err := validatePrescription(p.MedicationName, p.UnitsPerBox, p.DailyConsumption, p.BoxStartDate); err != nil {
		return Prescription{}, err
	}

	ok, err := s.deps.Consensus.HasConsensus(ctx, p.PatientID)
	if err != nil {
		return Prescription{}, fmt.Errorf("checking consensus: %w", err)
	}
	if !ok {
		return Prescription{}, ErrNoConsensus
	}

	rx, err := s.deps.Creator.Create(ctx, p)
	if err != nil {
		return Prescription{}, fmt.Errorf("creating prescription: %w", err)
	}
	return rx, nil
}

// Get returns a prescription by ID.
func (s *Service) Get(ctx context.Context, id int64) (Prescription, error) {
	return s.deps.Getter.GetByID(ctx, id)
}

// ListByPatient returns all prescriptions for a patient.
func (s *Service) ListByPatient(ctx context.Context, patientID int64) ([]Prescription, error) {
	return s.deps.Lister.ListByPatient(ctx, patientID)
}

// Update validates and updates a prescription.
func (s *Service) Update(ctx context.Context, p UpdateParams) error {
	if err := validatePrescription(p.MedicationName, p.UnitsPerBox, p.DailyConsumption, p.BoxStartDate); err != nil {
		return err
	}

	if err := s.deps.Updater.Update(ctx, p); err != nil {
		return fmt.Errorf("updating prescription: %w", err)
	}
	return nil
}

// RecordRefill delegates to the refill recorder.
func (s *Service) RecordRefill(ctx context.Context, p RefillParams) error {
	if err := s.deps.Refill.RecordRefill(ctx, p); err != nil {
		return fmt.Errorf("recording refill: %w", err)
	}
	return nil
}

func validatePrescription(medicationName string, unitsPerBox int, dailyConsumption float64, boxStartDate interface{ IsZero() bool }) error {
	if medicationName == "" {
		return errors.New("il nome del farmaco è obbligatorio")
	}
	if unitsPerBox <= 0 {
		return errors.New("le unità per confezione devono essere maggiori di zero")
	}
	if dailyConsumption <= 0 {
		return errors.New("il consumo giornaliero deve essere maggiore di zero")
	}
	if boxStartDate.IsZero() {
		return errors.New("la data di inizio confezione è obbligatoria")
	}
	return nil
}
