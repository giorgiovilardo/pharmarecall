package patient

import "context"

// PatientCreator creates a patient in a transaction.
type PatientCreator interface {
	Create(ctx context.Context, p CreateParams) (Patient, error)
}

// PatientGetter fetches a patient by ID.
type PatientGetter interface {
	GetByID(ctx context.Context, id int64) (Patient, error)
}

// PatientLister lists patients for a pharmacy.
type PatientLister interface {
	List(ctx context.Context, pharmacyID int64) ([]Summary, error)
}

// PatientUpdater updates a patient in a transaction.
type PatientUpdater interface {
	Update(ctx context.Context, p UpdateParams) error
}

// ConsensusRecorder records patient consensus.
type ConsensusRecorder interface {
	SetConsensus(ctx context.Context, id int64) error
}

// Repository composes all ports — used only by NewService for convenient wiring.
type Repository interface {
	PatientCreator
	PatientGetter
	PatientLister
	PatientUpdater
	ConsensusRecorder
}
