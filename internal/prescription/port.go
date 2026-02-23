package prescription

import "context"

// PrescriptionCreator creates a prescription in a transaction.
type PrescriptionCreator interface {
	Create(ctx context.Context, p CreateParams) (Prescription, error)
}

// PrescriptionGetter fetches a prescription by ID.
type PrescriptionGetter interface {
	GetByID(ctx context.Context, id int64) (Prescription, error)
}

// PrescriptionLister lists prescriptions for a patient.
type PrescriptionLister interface {
	ListByPatient(ctx context.Context, patientID int64) ([]Prescription, error)
}

// PrescriptionUpdater updates a prescription in a transaction.
type PrescriptionUpdater interface {
	Update(ctx context.Context, p UpdateParams) error
}

// RefillRecorder records a refill: updates box_start_date and inserts history.
type RefillRecorder interface {
	RecordRefill(ctx context.Context, p RefillParams) error
}

// Repository composes all ports — used only by NewService for convenient wiring.
type Repository interface {
	PrescriptionCreator
	PrescriptionGetter
	PrescriptionLister
	PrescriptionUpdater
	RefillRecorder
}
