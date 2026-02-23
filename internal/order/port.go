package order

import (
	"context"
	"time"
)

// OrderCreator creates a new order in a transaction.
type OrderCreator interface {
	Create(ctx context.Context, p CreateParams) (Order, error)
}

// ActiveOrderChecker checks if an active order exists for a prescription and cycle.
type ActiveOrderChecker interface {
	HasActiveOrder(ctx context.Context, prescriptionID int64, cycleStartDate time.Time) (bool, error)
}

// DashboardLister lists dashboard entries for a pharmacy.
type DashboardLister interface {
	ListDashboard(ctx context.Context, pharmacyID int64) ([]DashboardEntry, error)
}

// OrderStatusUpdater updates the status of an order in a transaction.
type OrderStatusUpdater interface {
	UpdateStatus(ctx context.Context, id int64, status string) error
}

// OrderGetter gets an order by ID.
type OrderGetter interface {
	GetByID(ctx context.Context, id int64) (Order, error)
}

// PrescriptionLookaheadLister lists prescriptions with consensus for a pharmacy.
type PrescriptionLookaheadLister interface {
	ListPrescriptionsForPharmacy(ctx context.Context, pharmacyID int64) ([]PrescriptionSummary, error)
}

// PrescriptionRefiller records a prescription refill when an order is fulfilled.
type PrescriptionRefiller interface {
	RecordRefill(ctx context.Context, prescriptionID int64, newStartDate time.Time) error
}

// Repository composes all ports — used only by NewService for convenient wiring.
type Repository interface {
	OrderCreator
	ActiveOrderChecker
	DashboardLister
	OrderStatusUpdater
	OrderGetter
	PrescriptionLookaheadLister
}
